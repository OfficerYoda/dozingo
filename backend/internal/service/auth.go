package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	authpkg "github.com/officeryoda/dozingo/internal/auth"
	"github.com/officeryoda/dozingo/internal/domain"
	"github.com/officeryoda/dozingo/internal/generated"
	"github.com/officeryoda/dozingo/internal/middleware"
	"github.com/officeryoda/dozingo/internal/repository"
)

type Auth struct {
	users              *repository.Users
	passwords          *repository.UserPasswords
	sessions           *repository.Sessions
	verificationTokens *repository.VerificationTokens
	queries            *generated.Queries
	txRunner           repository.TxRunner
}

func NewAuth(repos repository.Repos, queries *generated.Queries, txRunner repository.TxRunner) *Auth {
	return &Auth{
		users:              repos.Users,
		passwords:          repos.Passwords,
		sessions:           repos.Sessions,
		verificationTokens: repos.VerificationTokens,
		txRunner:           txRunner,
		queries:            queries,
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

type UpdatePasswordInput struct {
	Token       string
	NewPassword string
}

func (s *Auth) Register(ctx context.Context, in RegisterInput) (generated.User, error) {
	user, err := s.generateUser(ctx, in)
	if err != nil {
		return generated.User{}, fmt.Errorf("user creation: %w", err)
	}

	// Run session stuff outside the transaction so the user can recover
	// via login when something with the session fails
	err = s.attachUserToSession(ctx, user)
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
	err = s.attachUserToSession(ctx, vanillaUser)
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

	err = middleware.ClearSessionTokenCookie(ctx)
	if err != nil {
		slog.Warn("failed to clear session cookie on logout", "error", err)
	}

	return nil
}

func (s *Auth) Me(ctx context.Context) (generated.User, error) {
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

func (s *Auth) UpdatePassword(ctx context.Context, in UpdatePasswordInput) (generated.User, error) {
	token, err := s.verificationTokens.GetByToken(ctx, in.Token)
	if err != nil {
		return generated.User{}, fmt.Errorf("retrieve verification token: %w", err)
	}

	if token.Type != generated.TokenType("password_reset") {
		return generated.User{}, fmt.Errorf("invalid token type: %w", domain.ErrBadInput)
	}

	if token.ExpiresAt.Time.After(time.Now()) {
		return generated.User{}, fmt.Errorf("invalid expired: %w", domain.ErrGone)
	}

	passwordHash, err := authpkg.HashPassword(in.NewPassword)
	if err != nil {
		return generated.User{}, fmt.Errorf("hash password: %w", err)
	}

	err = s.txRunner.WithTx(ctx, func(r repository.Repos) error {
		err = r.VerificationTokens.Delete(ctx, token.Token)
		if err != nil {
			return fmt.Errorf("delete verification token : %w", err)
		}

		if err = r.Sessions.DeleteByUserID(ctx, token.UserID); err != nil {
			return fmt.Errorf("delete user sessions: %w", err)
		}

		if _, err = r.Passwords.Upsert(ctx, token.UserID, passwordHash); err != nil {
			return fmt.Errorf("update password: %w", err)
		}

		return nil
	})
	if err != nil {
		return generated.User{}, err
	}

	user, err := s.users.GetByID(ctx, token.UserID)
	if err != nil {
		return generated.User{}, fmt.Errorf("retrieve user: %w", err)
	}

	return user, nil
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

func (s *Auth) attachUserToSession(ctx context.Context, user generated.User) error {
	sessionUser, err := middleware.RequireSession(ctx, s.queries)
	if err != nil {
		return fmt.Errorf("session required: %w", err)
	}
	_, err = s.sessions.AttachUser(ctx, sessionUser.Token, user.ID)
	if err != nil {
		return fmt.Errorf("attach user to session: %w", err)
	}
	return nil
}
