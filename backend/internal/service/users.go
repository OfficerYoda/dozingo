package service

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/mail"
	"strings"

	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/officeryoda/dozingo/internal/auth"
	"github.com/officeryoda/dozingo/internal/domain"
	"github.com/officeryoda/dozingo/internal/email"
	"github.com/officeryoda/dozingo/internal/generated"
	"github.com/officeryoda/dozingo/internal/middleware"
	"github.com/officeryoda/dozingo/internal/pgmap"
	"github.com/officeryoda/dozingo/internal/repository"
	"github.com/officeryoda/dozingo/internal/storage"
)

// MaxAvatarBytes caps the size of a user-uploaded avatar payload
const MaxAvatarBytes = 20 * 1024 * 1024

var allowedAvatarMIMEs = map[string]string{
	"image/png":  ".png",
	"image/jpeg": ".jpg",
	"image/webp": ".webp",
}

type Users struct {
	users       *repository.Users
	passwords   *repository.UserPasswords
	queries     *generated.Queries
	emailSender email.Sender
	txRunner    repository.TxRunner
	uploader    storage.ObjectUploader
}

func NewUsers(
	repos *repository.Repos,
	queries *generated.Queries,
	emailSender email.Sender,
	txRunner repository.TxRunner,
	uploader storage.ObjectUploader,
) *Users {
	return &Users{
		users:       repos.Users,
		passwords:   repos.Passwords,
		queries:     queries,
		emailSender: emailSender,
		txRunner:    txRunner,
		uploader:    uploader,
	}
}

func (s *Users) Me(ctx context.Context) (generated.User, error) {
	sessionUser, err := requiresSessionUser(ctx, s.queries)
	if err != nil {
		return generated.User{}, err
	}

	return generated.User{
		ID:        sessionUser.UserID,
		Username:  sessionUser.Username.String,
		Email:     sessionUser.Email,
		AvatarKey: sessionUser.AvatarKey.String,
	}, nil
}

func (s *Users) Delete(ctx context.Context, password string) error {
	sessionUser, err := requiresSessionUser(ctx, s.queries)
	if err != nil {
		return err
	}

	err = s.txRunner.WithTx(ctx, func(r repository.Repos) error {
		var passwordHash string
		passwordHash, err = r.Passwords.GetHashForUserID(ctx, sessionUser.UserID)
		if err != nil {
			return fmt.Errorf("get password hash: %w", err)
		}

		err = auth.CheckPassword(password, passwordHash)
		if err != nil {
			return fmt.Errorf("password mismatch: %w", err)
		}

		_, err = r.Users.Delete(ctx, sessionUser.UserID)
		if err != nil {
			return fmt.Errorf("delete user: %w", err)
		}

		err = r.Sessions.DeleteByUserID(ctx, sessionUser.UserID)
		if err != nil {
			return fmt.Errorf("delete sessions: %w", err)
		}

		return nil
	})
	if err != nil {
		return err
	}

	return nil
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

	sessionUser, _ := middleware.SessionUserFromContext(ctx)
	if sessionUser.UserID != user.ID {
		user.Email = pgmap.PgTextFromString(nil)
	}

	return user, nil
}

// UpdateUserInput captures the PATCH payload for a user update:
//
//   - Username: nil -> leave alone, non-nil -> set to that value.
//   - Email: nil -> leave email and email_verified_at alone.
//     "" (empty string) -> clear email and reset email_verified_at to NULL.
//     non-empty string -> set email, reset email_verified_at to NULL,
//     and (if the address changed) dispatch a verification mail.
type UpdateUserInput struct {
	Username *string
	Email    *string
}

func (s *Users) UpdateMe(ctx context.Context, in UpdateUserInput) (generated.User, error) {
	sessionUser, err := requiresSessionUser(ctx, s.queries)
	if err != nil {
		return generated.User{}, err
	}

	return s.applyUserUpdate(ctx, sessionUser.UserID, sessionUser.Email, in)
}

func (s *Users) UploadAvatar(ctx context.Context, in huma.FormFile) (generated.User, error) {
	sessionUser, err := requiresSessionUser(ctx, s.queries)
	if err != nil {
		return generated.User{}, err
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
	if in.Size > MaxAvatarBytes {
		return nil, fmt.Errorf("avatar exceeds %d bytes: %w", MaxAvatarBytes, domain.ErrBadInput)
	}

	limited := io.LimitReader(in.File, MaxAvatarBytes+1)
	fileBytes, err := io.ReadAll(limited)
	if err != nil {
		return nil, fmt.Errorf("read avatar: %w", err)
	}
	if len(fileBytes) > MaxAvatarBytes {
		return nil, fmt.Errorf("avatar exceeds %d bytes: %w", MaxAvatarBytes, domain.ErrBadInput)
	}

	declared := strings.ToLower(strings.TrimSpace(in.ContentType))
	if declared == "image/svg+xml" || declared == "image/svg" {
		return nil, fmt.Errorf("svg avatars are not allowed: %w", domain.ErrBadInput)
	}

	sniffLen := min(len(fileBytes), 512)
	detected := http.DetectContentType(fileBytes[:sniffLen])
	// http.DetectContentType returns "type/subtype; charset=utf-8" sometimes
	if idx := strings.IndexByte(detected, ';'); idx >= 0 {
		detected = detected[:idx]
	}
	detected = strings.ToLower(strings.TrimSpace(detected))

	extension, ok := allowedAvatarMIMEs[detected]
	if !ok {
		return nil, fmt.Errorf("unsupported avatar type %q: %w", detected, domain.ErrBadInput)
	}

	bytesReader := bytes.NewReader(fileBytes)
	img := storage.NewImage(bytesReader, detected, extension)

	return img, nil
}

func (s *Users) applyUserUpdate(ctx context.Context, userID pgtype.UUID, prevEmail pgtype.Text, in UpdateUserInput) (generated.User, error) {
	if in.Email != nil && strings.TrimSpace(*in.Email) != "" {
		if _, err := mail.ParseAddress(*in.Email); err != nil {
			return generated.User{}, fmt.Errorf("invalid email address: %w", domain.ErrUnprocessableEntity)
		}
	}

	user, err := s.users.Update(ctx, userID, repository.UpdateUserParams{
		Username:   in.Username,
		ClearEmail: in.Email != nil && strings.TrimSpace(*in.Email) == "",
		Email:      in.Email,
	})
	if err != nil {
		return generated.User{}, fmt.Errorf("update user: %w", err)
	}

	if in.Email != nil && strings.TrimSpace(*in.Email) != "" && user.Email.Valid && !pgTextEqual(prevEmail, user.Email) {
		if err := sendEmailVerification(ctx, s.txRunner, s.emailSender, user.ID, user.Email.String); err != nil {
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
