package handler

import (
	"context"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/officeryoda/dozingo/internal/generated"
	"github.com/officeryoda/dozingo/internal/middleware"
	"github.com/officeryoda/dozingo/internal/pgmap"
	"github.com/officeryoda/dozingo/internal/service"
)

// ===== Input/Output types =====

type userOutputBody struct {
	UserID   string  `json:"user_id" format:"uuid"`
	Username string  `json:"username"`
	Email    *string `json:"email"`
}

type registerInputBody struct {
	Username string  `json:"username" required:"true" maxLength:"200"`
	Password string  `json:"password" required:"true" minLength:"8" maxLength:"72"`
	Email    *string `json:"email,omitempty" maxLength:"200"`
}

type registerInput struct {
	Body registerInputBody
}

type registerOutput struct {
	Body userOutputBody
}

type loginInputBody struct {
	Username string `json:"username" required:"true" maxLength:"200"`
	Password string `json:"password" required:"true" minLength:"8" maxLength:"72"`
}

type loginInput struct {
	Body loginInputBody
}

type loginOutput struct {
	Body userOutputBody
}

type meOutput struct {
	Body userOutputBody
}

// ===== Handler =====

type AuthHandler struct {
	svc *service.Auth
}

func NewAuthHandler(svc *service.Auth) *AuthHandler {
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

	huma.Register(api, huma.Operation{
		OperationID: "login",
		Method:      http.MethodPost,
		Path:        "/auth/login",
		Summary:     "Login with existing User",
		Tags:        []string{"Auth"},
	}, h.login)

	huma.Register(api, huma.Operation{
		OperationID: "logout",
		Method:      http.MethodPost,
		Path:        "/auth/logout",
		Summary:     "Logout from logged in User",
		Tags:        []string{"Auth"},
	}, h.logout)

	huma.Register(api, huma.Operation{
		OperationID: "me",
		Method:      http.MethodGet,
		Path:        "/auth/me",
		Summary:     "Information about the logged-in user",
		Tags:        []string{"Auth"},
	}, h.me)
}

func (h *AuthHandler) register(ctx context.Context, in *registerInput) (*registerOutput, error) {
	user, err := h.svc.Register(ctx, service.RegisterInput{
		Username: in.Body.Username,
		Password: in.Body.Password,
		Email:    in.Body.Email,
	})
	if err != nil {
		return nil, toHumaErr(err, "", "failed to register user")
	}

	return &registerOutput{Body: userToOutput(user)}, nil
}

func (h *AuthHandler) login(ctx context.Context, in *loginInput) (*loginOutput, error) {
	user, err := h.svc.Login(ctx, service.LoginInput{
		Username: in.Body.Username,
		Password: in.Body.Password,
	})
	if err != nil {
		return nil, toHumaErr(err, "", "failed to login user")
	}

	return &loginOutput{Body: userToOutput(user)}, nil
}

func (h *AuthHandler) logout(ctx context.Context, in *struct{}) (*struct{}, error) {
	err := h.svc.Logout(ctx)
	if err != nil {
		return nil, toHumaErr(err, "", "failed to logout user")
	}

	return &struct{}{}, nil
}

func (h *AuthHandler) me(ctx context.Context, in *struct{}) (*meOutput, error) {
	session, ok := middleware.SessionUserFromContext(ctx)
	if !ok || !session.UserID.Valid {
		return nil, huma.Error401Unauthorized("not logged in")
	}

	user, err := h.svc.Me(ctx, session)
	if err != nil {
		return nil, toHumaErr(err, "", "failed to get me")
	}

	return &meOutput{Body: userToOutput(user)}, nil
}

func userToOutput(user generated.User) userOutputBody {
	return userOutputBody{
		UserID:   user.ID.String(),
		Username: user.Username,
		Email:    pgmap.StringFromPgText(user.Email),
	}
}
