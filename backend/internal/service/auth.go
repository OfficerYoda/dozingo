package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	authpkg "github.com/officeryoda/dozingo/internal/auth"
	"github.com/officeryoda/dozingo/internal/domain"
	"github.com/officeryoda/dozingo/internal/generated"
	"github.com/officeryoda/dozingo/internal/middleware"
	"github.com/officeryoda/dozingo/internal/repository"
)

type Auth struct {
	users     *repository.Users
	passwords *repository.UserPasswords
	sessions  *repository.Sessions
	queries   *generated.Queries
	txRunner  repository.TxRunner
}

func NewAuth(repos repository.Repos, queries *generated.Queries, txRunner repository.TxRunner) *Auth {
	return &Auth{
		users:     repos.Users,
		passwords: repos.Passwords,
		sessions:  repos.Sessions,
		txRunner:  txRunner,
		queries:   queries,
	}
}

type RegisterInput struct {
	Username string
	Password string
	Email    *string
}

type LoginInput struct {
	Username string
	Password string
}

func (s *Auth) Register(ctx context.Context, in RegisterInput) (generated.User, error) {
	user, err := s.generateUser(ctx, in)
	if err != nil {
		return generated.User{}, fmt.Errorf("user creation: %w", err)
	}

	// Run session stuff outside the transaction so the user can recover
	// via login when something with the session fails
	err = s.attatchUserToSession(ctx, user)
	if err != nil {
		return generated.User{}, fmt.Errorf("attach user to session: %w", err)
	}

	return user, nil
}

func (s *Auth) Login(ctx context.Context, in LoginInput) (generated.User, error) {
	user, err := s.users.GetForPasswordLogin(ctx, in.Username)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			authpkg.CheckPasswordAgainstDummy(in.Password)
			return generated.User{}, domain.ErrUnauthorized
		}
		return generated.User{}, fmt.Errorf("user retrieval for login: %w", err)
	}

	err = authpkg.CheckPassword(in.Password, user.PasswordHash)
	if err != nil {
		return generated.User{}, domain.ErrUnauthorized
	}

	vanillaUser := generated.User{
		ID:       user.ID,
		Username: user.Username,
		Email:    user.Email,
	}
	err = s.attatchUserToSession(ctx, vanillaUser)
	if err != nil {
		return generated.User{}, fmt.Errorf("attach user to session: %w", err)
	}

	return vanillaUser, nil
}

func (s *Auth) Logout(ctx context.Context) error {
	sessionUser, ok := middleware.SessionUserFromContext(ctx)
	if !ok || !sessionUser.UserID.Valid {
		return nil
	}

	err := s.sessions.Delete(ctx, sessionUser.Token)
	if err != nil {
		return fmt.Errorf("delete session token: %w", err)
	}

	if err := middleware.ClearSessionTokenCookieCtx(ctx); err != nil {
		slog.Warn("failed to clear session cookie on logout", "error", err)
	}

	return nil
}

func (s *Auth) Me(ctx context.Context, session generated.GetSessionUserByTokenRow) (generated.User, error) {
	return generated.User{
		ID:       session.UserID,
		Username: session.Username.String,
		Email:    session.Email,
	}, nil
}

func (s *Auth) generateUser(ctx context.Context, in RegisterInput) (generated.User, error) {
	passwordHash, err := authpkg.HashPassword(in.Password)
	if err != nil {
		return generated.User{}, err
	}

	var user generated.User
	err = s.txRunner.WithTx(ctx, func(r repository.Repos) error {
		user, err = r.Users.Create(ctx, in.Username, in.Email)
		if err != nil {
			return err
		}
		_, err = r.Passwords.Upsert(ctx, user.ID, passwordHash)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return generated.User{}, err
	}
	return user, nil
}

func (s *Auth) attatchUserToSession(ctx context.Context, user generated.User) error {
	session, err := middleware.RequireSessionCtx(ctx, s.queries)
	if err != nil {
		return fmt.Errorf("require session: %w", err)
	}
	_, err = s.sessions.AttachUser(ctx, session.Token, user.ID)
	if err != nil {
		return fmt.Errorf("attach user to session: %w", err)
	}
	return nil
}
