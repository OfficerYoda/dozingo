package handler

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/officeryoda/dozingo/internal/generated"
	"github.com/officeryoda/dozingo/internal/middleware"
	"github.com/officeryoda/dozingo/internal/pgmap"
	"github.com/officeryoda/dozingo/internal/service"
)

// ===== Input/Output types =====

type userOutputBody struct {
	UserID   string  `json:"user_id"  format:"uuid"`
	Username string  `json:"username" pattern:"^[^\\s\\x00-\\x1F\\x7F]+$"`
	Email    *string `json:"email"    format:"email"`
}

type userOutput struct {
	Body userOutputBody
}

type registerInputBody struct {
	Username string  `json:"username" required:"true" maxLength:"200"`
	Password string  `json:"password" required:"true" minLength:"8" maxLength:"72"`
	Email    *string `json:"email,omitempty" format:"email" maxLength:"200"`
}

type registerInput struct {
	Body registerInputBody
}

type loginInputBody struct {
	Username string `json:"username" required:"true" maxLength:"200"`
	Password string `json:"password" required:"true" minLength:"8" maxLength:"72"`
}

type loginInput struct {
	Body loginInputBody
}

type forgotPasswordInputBody struct {
	Email string `json:"email" format:"email" required:"true" maxLength:"200"`
}

type forgotPasswordInput struct {
	Body forgotPasswordInputBody
}

type newPasswordInputBody struct {
	Token       string `json:"token" required:"true"`
	NewPassword string `json:"new_password" required:"true" minLength:"8" maxLength:"72"`
}

type newPasswordInput struct {
	Body newPasswordInputBody
}

type emailSentOutput struct {
	Body emailSentOutputBody
}

type emailSentOutputBody struct {
	Status string `json:"status"`
}

type verifyEmailInputBody struct {
	Token string `json:"token" required:"true"`
}

type verifyEmailInput struct {
	Body verifyEmailInputBody
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
		Middlewares: huma.Middlewares{middleware.RateLimit(api, middleware.StrictAuthLimiter)},
	}, h.register)

	huma.Register(api, huma.Operation{
		OperationID: "login",
		Method:      http.MethodPost,
		Path:        "/auth/login",
		Summary:     "Login with existing User",
		Tags:        []string{"Auth"},
		Middlewares: huma.Middlewares{middleware.RateLimit(api, middleware.HeavyAuthLimiter)},
	}, h.login)

	huma.Register(api, huma.Operation{
		OperationID: "logout",
		Method:      http.MethodPost,
		Path:        "/auth/logout",
		Summary:     "Logout from current User",
		Tags:        []string{"Auth"},
		Middlewares: huma.Middlewares{middleware.RateLimit(api, middleware.WriteLimiter)},
	}, h.logout)

	huma.Register(api, huma.Operation{
		OperationID: "forgot-password",
		Method:      http.MethodPost,
		Path:        "/auth/forgot-password",
		Summary:     "Request a password reset mail",
		Tags:        []string{"Auth"},
		Middlewares: huma.Middlewares{middleware.RateLimit(api, middleware.StrictAuthLimiter)},
	}, h.forgotPassword)

	huma.Register(api, huma.Operation{
		OperationID: "new-password",
		Method:      http.MethodPost,
		Path:        "/auth/new-password",
		Summary:     "Set a new password after reset",
		Tags:        []string{"Auth"},
		Middlewares: huma.Middlewares{middleware.RateLimit(api, middleware.StrictAuthLimiter)},
	}, h.newPassword)

	huma.Register(api, huma.Operation{
		OperationID: "send-email-verification",
		Method:      http.MethodPost,
		Path:        "/auth/send-email-verification",
		Summary:     "Send email verification mail",
		Tags:        []string{"Auth"},
		Middlewares: huma.Middlewares{middleware.RateLimit(api, middleware.StrictAuthLimiter)},
	}, h.sendEmailVerification)

	huma.Register(api, huma.Operation{
		OperationID: "verify-email",
		Method:      http.MethodPost,
		Path:        "/auth/verify-email",
		Summary:     "Verify an email",
		Tags:        []string{"Auth"},
		Middlewares: huma.Middlewares{middleware.RateLimit(api, middleware.StrictAuthLimiter)},
	}, h.verifyEmail)
}

func (h *AuthHandler) register(ctx context.Context, in *registerInput) (*userOutput, error) {
	user, err := h.svc.Register(ctx, service.RegisterInput{
		Username: in.Body.Username,
		Password: in.Body.Password,
		Email:    in.Body.Email,
	})
	if err != nil {
		return nil, toHumaErr(err, "", "failed to register user")
	}

	return &userOutput{Body: userToOutput(user)}, nil
}

func (h *AuthHandler) login(ctx context.Context, in *loginInput) (*userOutput, error) {
	user, err := h.svc.Login(ctx, service.LoginInput{
		Username: in.Body.Username,
		Password: in.Body.Password,
	})
	if err != nil {
		return nil, toHumaErr(err, "", "failed to login user")
	}

	return &userOutput{Body: userToOutput(user)}, nil
}

func (h *AuthHandler) logout(ctx context.Context, _ *struct{}) (*struct{}, error) {
	err := h.svc.Logout(ctx)
	if err != nil {
		return nil, toHumaErr(err, "", "failed to logout user")
	}

	return &struct{}{}, nil
}

func (h *AuthHandler) forgotPassword(ctx context.Context, in *forgotPasswordInput) (*emailSentOutput, error) {
	err := h.svc.ForgotPassword(ctx, in.Body.Email)
	if err != nil {
		slog.Warn("failed to send password reset email", "error", err)
		// swallow the error so attackers can't test for existing emails
	}

	return &emailSentOutput{Body: emailSentOutputBody{Status: "password reset email sent"}}, nil
}

func (h *AuthHandler) newPassword(ctx context.Context, in *newPasswordInput) (*userOutput, error) {
	user, err := h.svc.NewPassword(ctx, service.NewPasswordInput{
		Token:       in.Body.Token,
		NewPassword: in.Body.NewPassword,
	})
	if err != nil {
		return nil, toHumaErr(err, "", "failed to update password")
	}

	return &userOutput{Body: userToOutput(user)}, nil
}

func (h *AuthHandler) sendEmailVerification(ctx context.Context, _ *struct{}) (*emailSentOutput, error) {
	err := h.svc.SendEmailVerification(ctx)
	if err != nil {
		return nil, toHumaErr(err, "", "failed to send verification email")
	}

	return &emailSentOutput{Body: emailSentOutputBody{Status: "verification email sent"}}, nil
}

func (h *AuthHandler) verifyEmail(ctx context.Context, in *verifyEmailInput) (*userOutput, error) {
	user, err := h.svc.VerifyEmail(ctx, in.Body.Token)
	if err != nil {
		return nil, toHumaErr(err, "", "failed to verify email")
	}

	return &userOutput{Body: userToOutput(user)}, nil
}

func userToOutput(user generated.User) userOutputBody {
	return userOutputBody{
		UserID:   user.ID.String(),
		Username: user.Username,
		Email:    pgmap.StringFromPgText(user.Email),
	}
}
