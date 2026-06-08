package service

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"

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

// MaxAvatarBytes caps the size of a user-uploaded avatar payload. Bigger
// uploads are rejected before the file is buffered so a malicious client
// can't push the API process to memory exhaustion. Auto-generated avatars
// (used during register) bypass this path and aren't subject to the cap.
const MaxAvatarBytes = 20 * 1024 * 1024

// allowedAvatarMIMEs is the closed set of content-types that can come out
// of `http.DetectContentType` and still be accepted by the avatar upload
// endpoint. SVG is excluded on purpose: SVG can embed <script> tags, so
// serving user-supplied SVGs back to other users would be a stored-XSS
// vector. The mapped value is the canonical extension we use in the
// storage key, so the user-supplied filename never influences the key.
var allowedAvatarMIMEs = map[string]string{
	"image/png":  ".png",
	"image/jpeg": ".jpg",
	"image/webp": ".webp",
}

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

	uuid, err := uuid.NewRandom()
	if err != nil {
		return generated.User{}, fmt.Errorf("generate uuid: %w", err)
	}

	objectKey := fmt.Sprintf("%s%s", uuid, img.Extension)
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
	// First-pass guard: huma populates Size from the multipart header so
	// we can refuse oversized uploads before allocating a buffer for
	// them.
	if in.Size > MaxAvatarBytes {
		return nil, fmt.Errorf("avatar exceeds %d bytes: %w", MaxAvatarBytes, domain.ErrBadInput)
	}

	// Second-pass guard: read at most MaxAvatarBytes+1 bytes so an
	// under-declared Size can't sneak past. If we hit the +1 byte the
	// payload is over the cap and we reject.
	limited := io.LimitReader(in.File, MaxAvatarBytes+1)
	fileBytes, err := io.ReadAll(limited)
	if err != nil {
		return nil, fmt.Errorf("read avatar: %w", err)
	}
	if len(fileBytes) > MaxAvatarBytes {
		return nil, fmt.Errorf("avatar exceeds %d bytes: %w", MaxAvatarBytes, domain.ErrBadInput)
	}

	// Reject the declared content-type up front for the case we
	// explicitly do not want to host (SVG; XSS vector when served back
	// inline). The detected-MIME check below would also drop SVG since
	// it's not in the whitelist, but rejecting on the declared type
	// first gives a clearer error and avoids running the sniffer on a
	// known-bad payload.
	declared := strings.ToLower(strings.TrimSpace(in.ContentType))
	if declared == "image/svg+xml" || declared == "image/svg" {
		return nil, fmt.Errorf("svg avatars are not allowed: %w", domain.ErrBadInput)
	}

	// http.DetectContentType only looks at the first 512 bytes; pass it
	// the prefix and trust its decision over the (client-controlled)
	// declared content-type. The whitelist also dictates the extension
	// we pin to the storage key, so user-supplied filenames never
	// influence object naming.
	sniffLen := 512
	if len(fileBytes) < sniffLen {
		sniffLen = len(fileBytes)
	}
	detected := http.DetectContentType(fileBytes[:sniffLen])
	// http.DetectContentType returns "type/subtype; charset=utf-8" for
	// some formats; strip the parameters before lookup.
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
