package handler

import (
	"context"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/officeryoda/dozingo/internal/auth"
	"github.com/officeryoda/dozingo/internal/generated"
)

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
		return registerUser(ctx, queries, *input)
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

func registerUser(ctx context.Context, queries *generated.Queries, input RegisterInput) (*AuthOutput, error) {
	email := pgTextFromString(input.Body.Email)

	user, err := queries.CreateUser(ctx, generated.CreateUserParams{
		Username: input.Body.Username,
		Email:    email,
	})
	if err != nil {
		return nil, huma.Error500InternalServerError("failed to create user", err)
	}

	passwordHash, err := auth.HashPassword(input.Body.Password)
	if err != nil {
		return nil, huma.Error500InternalServerError("failed to hash password", err)
	}

	_, err = queries.UpsertUserPassword(ctx, generated.UpsertUserPasswordParams{
		UserID:       user.ID,
		PasswordHash: passwordHash,
	})
	if err != nil {
		return nil, huma.Error500InternalServerError("failed to create user password hash", err)
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
		auth.CheckPassword(input.Body.Password, "$2a$10$dummyHashToPreventTimingAttacksObiWanYouAreMyLastHope123")
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
