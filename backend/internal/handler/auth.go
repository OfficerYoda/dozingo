package handler

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/officeryoda/dozingo/internal/auth"
	"github.com/officeryoda/dozingo/internal/generated"
	"github.com/officeryoda/dozingo/internal/middleware"
)

/// ===== Input/Output types =====

type AuthOutput struct {
	Body struct {
		UserID   string  `json:"user_id" format:"uuid"`
		Username string  `json:"username" format:"text"`
		Email    *string `json:"email" format:"text"`
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
		return registerUser(ctx, pool, queries, *input)
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

	huma.Register(api, huma.Operation{
		OperationID: "logout",
		Method:      http.MethodPost,
		Path:        "/auth/logout",
		Summary:     "Logout from logged in User",
		Tags:        []string{"Auth"},
	}, func(ctx context.Context, input *struct{}) (*struct{}, error) {
		return logoutUser(ctx, queries)
	})

	huma.Register(api, huma.Operation{
		OperationID: "me",
		Method:      http.MethodGet,
		Path:        "/auth/me",
		Summary:     "Information about the logged-in user",
		Tags:        []string{"Auth"},
	}, func(ctx context.Context, input *struct{}) (*AuthOutput, error) {
		return getMe(ctx)
	})
}

/// ===== Handlers =====

func registerUser(ctx context.Context, pool *pgxpool.Pool, queries *generated.Queries, input RegisterInput) (*AuthOutput, error) {
	transaction, err := pool.Begin(ctx)
	if err != nil {
		return nil, internalError(err, "failed to create transaction")
	}
	defer func() {
		if err := transaction.Rollback(ctx); err != nil && !errors.Is(err, pgx.ErrTxClosed) {
			slog.Error("failed to rollback transaction", "error", err)
		}
	}()

	txQueries := generated.New(transaction)

	user, err := txQueries.CreateUser(ctx, generated.CreateUserParams{
		Username: input.Body.Username,
		Email:    pgTextFromString(input.Body.Email),
	})
	if err != nil {
		// check if it's a duplicate username/email error
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, huma.Error409Conflict("username or email already taken")
		}
		return nil, internalError(err, "failed to create user")
	}

	passwordHash, err := auth.HashPassword(input.Body.Password)
	if err != nil {
		return nil, internalError(err, "failed to hash password")
	}

	_, err = txQueries.UpsertUserPassword(ctx, generated.UpsertUserPasswordParams{
		UserID:       user.ID,
		PasswordHash: passwordHash,
	})
	if err != nil {
		return nil, internalError(err, "failed to create user password")
	}

	if err := transaction.Commit(ctx); err != nil {
		return nil, internalError(err, "failed to commit transaction")
	}

	// Session stuff runs against the non transaction pool, after the user is created
	// If anything below fails, the user still exists and can recover via login.
	session, err := middleware.RequireSessionCtx(ctx, queries)
	if err != nil {
		return nil, internalError(err, "failed to require session")
	}

	_, err = queries.AttachUserToSession(ctx, generated.AttachUserToSessionParams{
		Token:  session.Token,
		UserID: user.ID,
	})
	if err != nil {
		return nil, internalError(err, "failed to attach user to session")
	}

	output := &AuthOutput{}
	output.Body.UserID = user.ID.String()
	output.Body.Username = user.Username
	output.Body.Email = stringFromPgText(user.Email)
	return output, nil
}

func loginUser(ctx context.Context, queries *generated.Queries, input LoginInput) (*AuthOutput, error) {
	user, err := queries.GetUserForPasswordLogin(ctx, input.Body.Username)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			auth.CheckPasswordAgainstDummy(input.Body.Password)
			return nil, huma.Error401Unauthorized("invalid credentials")
		}
		return nil, internalError(err, "failed to fetch user for login")
	}

	err = auth.CheckPassword(input.Body.Password, user.PasswordHash)
	if err != nil {
		return nil, huma.Error401Unauthorized("invalid credentials")
	}

	session, err := middleware.RequireSessionCtx(ctx, queries)
	if err != nil {
		return nil, internalError(err, "failed to require session")
	}

	_, err = queries.AttachUserToSession(ctx, generated.AttachUserToSessionParams{
		Token:  session.Token,
		UserID: user.ID,
	})
	if err != nil {
		return nil, internalError(err, "failed to attach user to session")
	}

	output := &AuthOutput{}
	output.Body.UserID = user.ID.String()
	output.Body.Username = user.Username
	output.Body.Email = stringFromPgText(user.Email)
	return output, nil
}

func logoutUser(ctx context.Context, queries *generated.Queries) (*struct{}, error) {
	session, ok := middleware.SessionUserFromContext(ctx)
	if !ok || !session.UserID.Valid {
		// nothing to log out on an anonymous user
		return &struct{}{}, nil
	}

	if err := queries.DeleteSessionByToken(ctx, session.Token); err != nil {
		slog.Warn("failed to delete session on logout", "error", err)
	}
	if err := middleware.ClearSessionTokenCookieCtx(ctx); err != nil {
		slog.Warn("failed to clear session cookie on logout", "error", err)
	}
	return &struct{}{}, nil
}

func getMe(ctx context.Context) (*AuthOutput, error) {
	session, ok := middleware.SessionUserFromContext(ctx)
	if !ok || !session.UserID.Valid {
		return nil, huma.Error401Unauthorized("not logged in")
	}

	output := &AuthOutput{}
	output.Body.UserID = session.UserID.String()
	output.Body.Username = session.Username.String
	output.Body.Email = stringFromPgText(session.Email)
	return output, nil
}
