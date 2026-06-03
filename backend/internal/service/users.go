package service

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/officeryoda/dozingo/internal/domain"
	"github.com/officeryoda/dozingo/internal/email"
	"github.com/officeryoda/dozingo/internal/generated"
	"github.com/officeryoda/dozingo/internal/middleware"
	"github.com/officeryoda/dozingo/internal/pgmap"
	"github.com/officeryoda/dozingo/internal/repository"
)

type Users struct {
	users       *repository.Users
	queries     *generated.Queries
	emailSender email.Sender
	txRunner    repository.TxRunner
}

func NewUsers(repos repository.Repos, queries *generated.Queries, emailSender email.Sender, txRunner repository.TxRunner) *Users {
	return &Users{
		users:       repos.Users,
		queries:     queries,
		emailSender: emailSender,
		txRunner:    txRunner,
	}
}

// Me returns the user backing the current session, or ErrUnauthorized if the
// caller is anonymous / unauthenticated. The data is read off the session row
// so we avoid an extra DB hit.
func (s *Users) Me(ctx context.Context) (generated.User, error) {
	session, ok := middleware.SessionUserFromContext(ctx)
	if !ok || !session.UserID.Valid {
		return generated.User{}, fmt.Errorf("not logged in: %w", domain.ErrUnauthorized)
	}

	return generated.User{
		ID:       session.UserID,
		Username: session.Username.String,
		Email:    session.Email,
	}, nil
}

// UserByID looks up a user by its UUID string. An unparseable UUID surfaces as
// ErrBadInput; a missing row as ErrNotFound (translated by the repo layer).
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

// UpdateUser applies a partial update to a user. Only the user themselves may
// edit their own row (self-only). Username collisions surface as ErrConflict
// via the unique-index translation in the repo layer.
func (s *Users) UpdateUser(ctx context.Context, userIDStr string, in UpdateUserInput) (generated.User, error) {
	userID := pgmap.PgUUIDFromString(&userIDStr)
	if !userID.Valid {
		return generated.User{}, fmt.Errorf("invalid UUID: %w", domain.ErrBadInput)
	}

	sessionUser, err := requiresSessionUser(ctx, s.queries)
	if err != nil {
		return generated.User{}, fmt.Errorf("require session: %w", err)
	}

	if sessionUser.UserID.Bytes != userID.Bytes {
		return generated.User{}, fmt.Errorf("cannot edit another user: %w", domain.ErrForbidden)
	}

	// Capture the previous email so we only re-send verification when the
	// address actually changed.
	prevEmail := sessionUser.Email

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

// pgTextEqual compares two pgtype.Text values: equal iff both are NULL, or
// both Valid and have the same string content.
func pgTextEqual(a, b pgtype.Text) bool {
	if a.Valid != b.Valid {
		return false
	}
	if !a.Valid {
		return true
	}
	return a.String == b.String
}

// requiresSessionUser is the shared "must be a logged-in user" guard for
// service methods. Used by both Auth and Users.
func requiresSessionUser(ctx context.Context, queries *generated.Queries) (generated.GetSessionUserByTokenRow, error) {
	sessionUser, err := middleware.RequireSession(ctx, queries)
	if err != nil {
		return generated.GetSessionUserByTokenRow{}, fmt.Errorf("session required: %w", err)
	}
	if !sessionUser.UserID.Valid {
		return generated.GetSessionUserByTokenRow{}, fmt.Errorf("authenticated user required: %w", domain.ErrUnauthorized)
	}
	return sessionUser, nil
}
