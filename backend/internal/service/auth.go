package service

import (
	"context"
	"fmt"

	authpkg "github.com/officeryoda/dozingo/internal/auth"
	"github.com/officeryoda/dozingo/internal/generated"
	"github.com/officeryoda/dozingo/internal/middleware"
	"github.com/officeryoda/dozingo/internal/repository"
)

type Auth interface {
	Register(ctx context.Context, in RegisterInput) (generated.User, error)
	Login(ctx context.Context, in LoginInput) (generated.User, error)
	Logout(ctx context.Context, session generated.GetSessionUserByTokenRow) (generated.User, error)
	Me(ctx context.Context, session generated.GetSessionUserByTokenRow) (generated.User, error)
}

type auth struct {
	users     *repository.Users
	passwords *repository.UserPasswords
	sessions  *repository.Sessions
	txRunner  repository.TxRunner
	queries   *generated.Queries
}

func NewAuth(repo repository.Repos, txRunner repository.TxRunner, queries *generated.Queries) Auth {
	return &auth{
		users:     repo.Users,
		passwords: repo.Passwords,
		sessions:  repo.Sessions,
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

// Register implements [Auth].
func (s *auth) Register(ctx context.Context, in RegisterInput) (generated.User, error) {
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

// Login implements [Auth].
func (s *auth) Login(ctx context.Context, in LoginInput) (generated.User, error) {
	panic("unimplemented")
}

// Logout implements [Auth].
func (s *auth) Logout(ctx context.Context, session generated.GetSessionUserByTokenRow) (generated.User, error) {
	panic("unimplemented")
}

// Me implements [Auth].
func (s *auth) Me(ctx context.Context, session generated.GetSessionUserByTokenRow) (generated.User, error) {
	panic("unimplemented")
}

func (s *auth) generateUser(ctx context.Context, in RegisterInput) (generated.User, error) {
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

func (s *auth) attatchUserToSession(ctx context.Context, user generated.User) error {
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

var _ Auth = (*auth)(nil)
