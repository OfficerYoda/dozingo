package service

import (
	"context"
	"fmt"

	"github.com/officeryoda/dozingo/internal/domain"
	"github.com/officeryoda/dozingo/internal/generated"
	"github.com/officeryoda/dozingo/internal/middleware"
	"github.com/officeryoda/dozingo/internal/pgmap"
	"github.com/officeryoda/dozingo/internal/repository"
)

type Users struct {
	users   *repository.Users
	queries *generated.Queries
}

func NewUsers(repos repository.Repos, queries *generated.Queries) *Users {
	return &Users{
		users:   repos.Users,
		queries: queries,
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
