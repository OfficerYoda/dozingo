// Package handler exposes the Huma HTTP handlers that translate API requests into service calls.
package handler

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/danielgtaylor/huma/v2"

	"github.com/officeryoda/dozingo/internal/avatar"
	"github.com/officeryoda/dozingo/internal/generated"
	"github.com/officeryoda/dozingo/internal/middleware"
	"github.com/officeryoda/dozingo/internal/service"
)

// ===== Input/Output types =====

type userOutputBody struct {
	UserID    string `json:"user_id" format:"uuid"`
	Username  string `json:"username" pattern:"^[^\\s\\x00-\\x1F\\x7F]+$"`
	Email     string `json:"email,omitempty" format:"email"`
	AvatarURL string `json:"avatar_url" format:"uri"`
}

type userOutput struct {
	Body userOutputBody
}

// loginOutputBody is the response body for POST /auth/login.
// When TwoFAPending is true the session has been marked pending and the
// user must call POST /auth/2fa/verify or /auth/2fa/verify-recovery before
// accessing protected endpoints. The user fields are empty in that case.
type loginOutputBody struct {
	TwoFAPending bool   `json:"two_fa_pending"`
	UserID       string `json:"user_id,omitempty"    format:"uuid"`
	Username     string `json:"username,omitempty"   pattern:"^[^\\s\\x00-\\x1F\\x7F]+$"`
	Email        string `json:"email,omitempty"      format:"email"`
	AvatarURL    string `json:"avatar_url,omitempty" format:"uri"`
}

type loginOutput struct {
	Body loginOutputBody
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
	Identifier string `json:"identifier" required:"true" maxLength:"200"`
	Password   string `json:"password" required:"true" minLength:"8" maxLength:"72"`
}

type loginInput struct {
	Body loginInputBody
}

type changePasswordInputBody struct {
	OldPassword string `json:"old_password" required:"true" minLength:"8" maxLength:"72"`
	NewPassword string `json:"new_password" required:"true" minLength:"8" maxLength:"72"`
}

type changePasswordInput struct {
	Body changePasswordInputBody
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
	svc               *service.Auth
	avatarURLs        *avatar.URLBuilder
	fallbackAvatarURL string
}

func NewAuthHandler(svc *service.Auth, avatarURLs *avatar.URLBuilder, fallbackAvatarURL string) *AuthHandler {
	return &AuthHandler{
		svc:               svc,
		avatarURLs:        avatarURLs,
		fallbackAvatarURL: fallbackAvatarURL,
	}
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
		OperationID: "change-password",
		Method:      http.MethodPost,
		Path:        "/auth/change-password",
		Summary:     "Change the current password",
		Tags:        []string{"Auth"},
		Middlewares: huma.Middlewares{middleware.RateLimit(api, middleware.StrictAuthLimiter)},
	}, h.changePassword)

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

	return &userOutput{Body: userToOutput(user, h.avatarURLs, h.fallbackAvatarURL)}, nil
}

func (h *AuthHandler) login(ctx context.Context, in *loginInput) (*loginOutput, error) {
	result, err := h.svc.Login(ctx, service.LoginInput{
		Identifier: in.Body.Identifier,
		Password:   in.Body.Password,
	})
	if err != nil {
		return nil, toHumaErr(err, "", "failed to login user")
	}

	if result.TwoFAPending {
		return &loginOutput{Body: loginOutputBody{TwoFAPending: true}}, nil
	}

	u := userToOutput(result.User, h.avatarURLs, h.fallbackAvatarURL)

	return &loginOutput{Body: loginOutputBody{
		TwoFAPending: false,
		UserID:       u.UserID,
		Username:     u.Username,
		Email:        u.Email,
		AvatarURL:    u.AvatarURL,
	}}, nil
}

func (h *AuthHandler) logout(ctx context.Context, _ *struct{}) (*struct{}, error) {
	err := h.svc.Logout(ctx)
	if err != nil {
		return nil, toHumaErr(err, "", "failed to logout user")
	}

	return &struct{}{}, nil
}

func (h *AuthHandler) changePassword(ctx context.Context, in *changePasswordInput) (*struct{}, error) {
	err := h.svc.ChangePassword(ctx, in.Body.OldPassword, in.Body.NewPassword)
	if err != nil {
		return nil, toHumaErr(err, "", "failed to update password")
	}

	return nil, nil
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

	return &userOutput{Body: userToOutput(user, h.avatarURLs, h.fallbackAvatarURL)}, nil
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

	return &userOutput{Body: userToOutput(user, h.avatarURLs, h.fallbackAvatarURL)}, nil
}

func userToOutput(user generated.User, avatarURLs *avatar.URLBuilder, fallbackURL string) userOutputBody {
	return userOutputBody{
		UserID:    user.ID.String(),
		Username:  user.Username,
		Email:     user.Email,
		AvatarURL: avatarURLs.URL(user.AvatarKey, fallbackURL),
	}
}
