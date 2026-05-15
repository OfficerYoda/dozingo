package handler

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/officeryoda/dozingo/internal/auth"
	"github.com/officeryoda/dozingo/internal/generated"
)

/// ===== Dummy hash for timing attack prevention =====

var dummyPasswordHash string

func init() {
	h, err := auth.HashPassword("dummy-password")
	if err != nil {
		panic("failed to generate dummy hash")
	}
	dummyPasswordHash = h
}

/// ===== Input/Output types =====

type AuthOutput struct {
	Body struct {
		Username string  `json:"username" format:"text" maxLength:"200"`
		Email    *string `json:"email" format:"text" maxLength:"200"`
	}
}

type RegisterInput struct {
	Body struct {
		Username string  `json:"username" format:"text" required:"true" maxLength:"200"`
		Password string  `json:"password" format:"text" required:"true" maxLength:"72"`
		Email    *string `json:"email,omitempty" format:"text" maxLength:"200"`
	}
}

type LoginInput struct {
	Body struct {
		Username string `json:"username" format:"text" required:"true" maxLength:"200"`
		Password string `json:"password" format:"text" required:"true" maxLength:"72"`
	}
}

/// ===== Register =====

func RegisterAuth(api huma.API, pool *pgxpool.Pool) {
	queries := generated.New(pool)

	huma.Register(api, huma.Operation{
		OperationID: "register",
		Method:      http.MethodPost,
		Path:        "/auth/register",
		Summary:     "Register new User",
		Tags:        []string{"Auth"},
	}, func(ctx context.Context, input *RegisterInput) (*AuthOutput, error) {
		return registerUser(ctx, pool, *input)
	})

	huma.Register(api, huma.Operation{
		OperationID: "login",
		Method:      http.MethodPost,
		Path:        "/auth/login",
		Summary:     "Login with existing User",
		Tags:        []string{"Auth"},
	}, func(ctx context.Context, input *LoginInput) (*AuthOutput, error) {
		return loginUser(ctx, queries, *input)
	})
}

/// ===== Handlers =====

func registerUser(ctx context.Context, pool *pgxpool.Pool, input RegisterInput) (*AuthOutput, error) {
	tx, err := pool.Begin(ctx)
	if err != nil {
		slog.Error("failed to create transaction", "error", err)
		return nil, huma.Error500InternalServerError("internal error")
	}
	defer func() {
		if err := tx.Rollback(ctx); err != nil {
			slog.Error("failed to rollback transaction", "error", err)
		}
	}()

	queries := generated.New(tx)

	user, err := queries.CreateUser(ctx, generated.CreateUserParams{
		Username: input.Body.Username,
		Email:    pgTextFromString(input.Body.Email),
	})
	if err != nil {
		// check if it's a duplicate username/email error
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, huma.Error409Conflict("username or email already taken")
		}
		slog.Error("failed to create user", "error", err)
		return nil, huma.Error500InternalServerError("internal error")
	}

	passwordHash, err := auth.HashPassword(input.Body.Password)
	if err != nil {
		slog.Error("failed to hash password", "error", err)
		return nil, huma.Error500InternalServerError("internal error")
	}

	_, err = queries.UpsertUserPassword(ctx, generated.UpsertUserPasswordParams{
		UserID:       user.ID,
		PasswordHash: passwordHash,
	})
	if err != nil {
		slog.Error("failed to create user password", "error", err)
		return nil, huma.Error500InternalServerError("internal error")
	}

	err = tx.Commit(ctx)
	if err != nil {
		slog.Error("failed to commit transaction", "error", err)
		return nil, huma.Error500InternalServerError("internal error")
	}

	output := &AuthOutput{}
	output.Body.Username = user.Username
	output.Body.Email = stringFromPgText(user.Email)
	return output, nil
}

func loginUser(ctx context.Context, queries *generated.Queries, input LoginInput) (*AuthOutput, error) {
	user, err := queries.GetUserForPasswordLogin(ctx, input.Body.Username)
	if err != nil {
		// Check password against dummy hash to prevent timing attacks
		_ = auth.CheckPassword(input.Body.Password, dummyPasswordHash)
		return nil, huma.Error401Unauthorized("invalid credentials")
	}

	err = auth.CheckPassword(input.Body.Password, user.PasswordHash)
	if err != nil {
		return nil, huma.Error401Unauthorized("invalid credentials")
	}

	output := &AuthOutput{}
	output.Body.Username = user.Username
	output.Body.Email = stringFromPgText(user.Email)
	return output, nil
}
