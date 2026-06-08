package service

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"path/filepath"

	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/officeryoda/dozingo/internal/domain"
	"github.com/officeryoda/dozingo/internal/email"
	"github.com/officeryoda/dozingo/internal/generated"
	"github.com/officeryoda/dozingo/internal/middleware"
	"github.com/officeryoda/dozingo/internal/pgmap"
	"github.com/officeryoda/dozingo/internal/repository"
	"github.com/officeryoda/dozingo/internal/storage"
)

type Users struct {
	users       *repository.Users
	queries     *generated.Queries
	emailSender email.Sender
	txRunner    repository.TxRunner
	uploader    storage.ObjectUploader
}

func NewUsers(
	repos repository.Repos,
	queries *generated.Queries,
	emailSender email.Sender,
	txRunner repository.TxRunner,
	uploader storage.ObjectUploader,
) *Users {
	return &Users{
		users:       repos.Users,
		queries:     queries,
		emailSender: emailSender,
		txRunner:    txRunner,
		uploader:    uploader,
	}
}

func (s *Users) Me(ctx context.Context) (generated.User, error) {
	session, ok := middleware.SessionUserFromContext(ctx)
	if !ok || !session.UserID.Valid {
		return generated.User{}, fmt.Errorf("not logged in: %w", domain.ErrUnauthorized)
	}

	return generated.User{
		ID:        session.UserID,
		Username:  session.Username.String,
		Email:     session.Email,
		AvatarKey: session.AvatarKey.String,
	}, nil
}

func (s *Users) UserByID(ctx context.Context, userIDStr string) (generated.User, error) {
	userID := pgmap.PgUUIDFromString(&userIDStr)
	if !userID.Valid {
		return generated.User{}, fmt.Errorf("invalid UUID: %w", domain.ErrBadInput)
	}

	user, err := s.users.GetByID(ctx, userID)
	if err != nil {
		return generated.User{}, fmt.Errorf("get user by id: %w", err)
	}

	return user, nil
}

// UpdateUserInput captures a tri-state PATCH payload:
//
//   - Username: nil -> leave alone, non-nil -> set to that value.
//   - EmailSet: false -> leave Email and email_verified_at alone.
//     true -> write Email (which may itself be nil to clear the column)
//     and reset email_verified_at to NULL.
//
// When EmailSet is true and the new Email differs from the previous value,
// the service fires a verification mail to the new address (when non-nil).
type UpdateUserInput struct {
	Username *string
	EmailSet bool
	Email    *string
}

func (s *Users) UpdateUser(ctx context.Context, userIDStr string, in UpdateUserInput) (generated.User, error) {
	userID := pgmap.PgUUIDFromString(&userIDStr)
	if !userID.Valid {
		return generated.User{}, fmt.Errorf("invalid UUID: %w", domain.ErrBadInput)
	}

	sessionUser, err := requiresSessionUser(ctx, s.queries)
	if err != nil {
		return generated.User{}, fmt.Errorf("session required: %w", err)
	}

	if sessionUser.UserID.Bytes != userID.Bytes {
		return generated.User{}, fmt.Errorf("cannot edit another user: %w", domain.ErrForbidden)
	}

	return s.applyUserUpdate(ctx, userID, sessionUser.Email, in)
}

func (s *Users) UpdateMe(ctx context.Context, in UpdateUserInput) (generated.User, error) {
	sessionUser, err := requiresSessionUser(ctx, s.queries)
	if err != nil {
		return generated.User{}, fmt.Errorf("require session: %w", err)
	}

	return s.applyUserUpdate(ctx, sessionUser.UserID, sessionUser.Email, in)
}

func (s *Users) UploadAvatar(ctx context.Context, in huma.FormFile) (generated.User, error) {
	sessionUser, err := requiresSessionUser(ctx, s.queries)
	if err != nil {
		return generated.User{}, fmt.Errorf("require session: %w", err)
	}

	img, err := convertFormFileToImage(in)
	if err != nil {
		return generated.User{}, err
	}

	id, err := uuid.NewRandom()
	if err != nil {
		return generated.User{}, fmt.Errorf("generate uuid: %w", err)
	}

	objectKey := fmt.Sprintf("%s%s", id, img.Extension)
	err = s.uploader.Upload(ctx, objectKey, img)
	if err != nil {
		return generated.User{}, fmt.Errorf("upload avatar: %w", err)
	}

	user, err := s.users.SetAvatar(ctx, sessionUser.UserID, objectKey)
	if err != nil {
		return generated.User{}, fmt.Errorf("set avatar key: %w", err)
	}

	return user, nil
}

func convertFormFileToImage(in huma.FormFile) (*storage.Image, error) {
	extension := filepath.Ext(in.Filename)

	fileBytes, err := io.ReadAll(in.File)
	if err != nil {
		return nil, fmt.Errorf("read file: %w", err)
	}
	bytesReader := bytes.NewReader(fileBytes)

	img := storage.NewImage(bytesReader, in.ContentType, extension)

	return img, nil
}

func (s *Users) applyUserUpdate(ctx context.Context, userID pgtype.UUID, prevEmail pgtype.Text, in UpdateUserInput) (generated.User, error) {
	user, err := s.users.Update(ctx, userID, repository.UpdateUserParams{
		Username: in.Username,
		EmailSet: in.EmailSet,
		Email:    in.Email,
	})
	if err != nil {
		return generated.User{}, fmt.Errorf("update user: %w", err)
	}

	if in.EmailSet && user.Email.Valid && !pgTextEqual(prevEmail, user.Email) {
		if err := issueAndSendEmailVerification(ctx, s.txRunner, s.emailSender, user.ID, user.Email.String); err != nil {
			return generated.User{}, fmt.Errorf("send verification mail: %w", err)
		}
	}

	return user, nil
}

func pgTextEqual(a, b pgtype.Text) bool {
	if a.Valid != b.Valid {
		return false
	}
	if !a.Valid {
		return true
	}

	return a.String == b.String
}

func requiresSessionUser(ctx context.Context, queries *generated.Queries) (generated.GetSessionUserByTokenRow, error) {
	sessionUser, err := middleware.RequireSession(ctx, queries)
	if err != nil {
		return generated.GetSessionUserByTokenRow{}, fmt.Errorf("session required: %w", err)
	}
	if !sessionUser.UserID.Valid {
		return generated.GetSessionUserByTokenRow{}, fmt.Errorf("requires authenticated user: %w", domain.ErrUnauthorized)
	}

	return sessionUser, nil
}
