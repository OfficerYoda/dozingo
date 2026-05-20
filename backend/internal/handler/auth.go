package handler

import (
	"context"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/officeryoda/dozingo/internal/generated"
	"github.com/officeryoda/dozingo/internal/service"
)

/// ===== Input/Output types =====

type authOutput struct {
	Body authOutputBody
}

type authOutputBody struct {
	UserID   string  `json:"user_id" format:"uuid"`
	Username string  `json:"username" format:"text"`
	Email    *string `json:"email" format:"text"`
}

type registerInput struct {
	Body struct {
		Username string  `json:"username" format:"text" required:"true" maxLength:"200"`
		Password string  `json:"password" format:"text" required:"true" maxLength:"72"`
		Email    *string `json:"email,omitempty" format:"text" maxLength:"200"`
	}
}

type loginInput struct {
	Body struct {
		Username string `json:"username" format:"text" required:"true" maxLength:"200"`
		Password string `json:"password" format:"text" required:"true" maxLength:"72"`
	}
}

/// ===== Handler =====

type AuthHandler struct {
	svc service.Auth
}

func NewAuthHandler(svc service.Auth) *AuthHandler {
	return &AuthHandler{svc: svc}
}

func (h *AuthHandler) Register(api huma.API) {
	huma.Register(api, huma.Operation{
		OperationID: "register",
		Method:      http.MethodPost,
		Path:        "/auth/register",
		Summary:     "Register new User",
		Tags:        []string{"Auth"},
	}, h.register)

	// queries := generated.New(pool)
	//
	// huma.Register(api, huma.Operation{
	// 	OperationID: "login",
	// 	Method:      http.MethodPost,
	// 	Path:        "/auth/login",
	// 	Summary:     "Login with existing User",
	// 	Tags:        []string{"Auth"},
	// }, func(ctx context.Context, input *LoginInput) (*AuthOutput, error) {
	// 	return loginUser(ctx, queries, *input)
	// })
	//
	// huma.Register(api, huma.Operation{
	// 	OperationID: "logout",
	// 	Method:      http.MethodPost,
	// 	Path:        "/auth/logout",
	// 	Summary:     "Logout from logged in User",
	// 	Tags:        []string{"Auth"},
	// }, func(ctx context.Context, input *struct{}) (*struct{}, error) {
	// 	return logoutUser(ctx, queries)
	// })
	//
	// huma.Register(api, huma.Operation{
	// 	OperationID: "me",
	// 	Method:      http.MethodGet,
	// 	Path:        "/auth/me",
	// 	Summary:     "Information about the logged-in user",
	// 	Tags:        []string{"Auth"},
	// }, func(ctx context.Context, input *struct{}) (*AuthOutput, error) {
	// 	return getMe(ctx)
	// })
}

/// ===== Handlers =====

func (h *AuthHandler) register(ctx context.Context, in *registerInput) (*authOutput, error) {
	user, err := h.svc.Register(ctx, service.RegisterInput{
		Username: in.Body.Username,
		Password: in.Body.Password,
		Email:    in.Body.Email,
	})
	if err != nil {
		return nil, toHumaErr(err, "", "failed to register user")
	}

	return &authOutput{Body: userToOutput(user)}, nil
}

// func loginUser(ctx context.Context, queries *generated.Queries, input LoginInput) (*AuthOutput, error) {
// 	user, err := queries.GetUserForPasswordLogin(ctx, input.Body.Username)
// 	if err != nil {
// 		if errors.Is(err, pgx.ErrNoRows) {
// 			auth.CheckPasswordAgainstDummy(input.Body.Password)
// 			return nil, huma.Error401Unauthorized("invalid credentials")
// 		}
// 		return nil, internalError(err, "failed to fetch user for login")
// 	}
//
// 	err = auth.CheckPassword(input.Body.Password, user.PasswordHash)
// 	if err != nil {
// 		return nil, huma.Error401Unauthorized("invalid credentials")
// 	}
//
// 	session, err := middleware.RequireSessionCtx(ctx, queries)
// 	if err != nil {
// 		return nil, internalError(err, "failed to require session")
// 	}
//
// 	_, err = queries.AttachUserToSession(ctx, generated.AttachUserToSessionParams{
// 		Token:  session.Token,
// 		UserID: user.ID,
// 	})
// 	if err != nil {
// 		return nil, internalError(err, "failed to attach user to session")
// 	}
//
// 	output := &AuthOutput{}
// 	output.Body.UserID = user.ID.String()
// 	output.Body.Username = user.Username
// 	output.Body.Email = stringFromPgText(user.Email)
// 	return output, nil
// }
//
// func logoutUser(ctx context.Context, queries *generated.Queries) (*struct{}, error) {
// 	session, ok := middleware.SessionUserFromContext(ctx)
// 	if !ok || !session.UserID.Valid {
// 		// nothing to log out on an anonymous user
// 		return &struct{}{}, nil
// 	}
//
// 	if err := queries.DeleteSessionByToken(ctx, session.Token); err != nil {
// 		slog.Warn("failed to delete session on logout", "error", err)
// 	}
// 	if err := middleware.ClearSessionTokenCookieCtx(ctx); err != nil {
// 		slog.Warn("failed to clear session cookie on logout", "error", err)
// 	}
// 	return &struct{}{}, nil
// }
//
// func getMe(ctx context.Context) (*AuthOutput, error) {
// 	session, ok := middleware.SessionUserFromContext(ctx)
// 	if !ok || !session.UserID.Valid {
// 		return nil, huma.Error401Unauthorized("not logged in")
// 	}
//
// 	output := &AuthOutput{}
// 	output.Body.UserID = session.UserID.String()
// 	output.Body.Username = session.Username.String
// 	output.Body.Email = stringFromPgText(session.Email)
// 	return output, nil
// }

func userToOutput(user generated.User) authOutputBody {
	return authOutputBody{
		UserID:   user.ID.String(),
		Username: user.Username,
		Email:    stringFromPgText(user.Email),
	}
}
