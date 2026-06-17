package handler

import (
	"context"
	"net/http"

	"github.com/danielgtaylor/huma/v2"

	"github.com/officeryoda/dozingo/internal/middleware"
	"github.com/officeryoda/dozingo/internal/service"
)

// ===== Input/Output types =====

type twoFASetupOutputBody struct {
	Secret     string `json:"secret"`
	OTPAuthURL string `json:"otp_auth_url"`
}

type twoFASetupOutput struct {
	Body twoFASetupOutputBody
}

type twoFACodeInput struct {
	Body struct {
		Code string `json:"code" pattern:"^\\d{6}$" minLength:"6" maxLength:"6" example:"123456"`
	}
}

type twoFAConfirmOutputBody struct {
	RecoveryCodes []string `json:"recovery_codes"`
}

type twoFAConfirmOutput struct {
	Body twoFAConfirmOutputBody
}

type twoFARecoveryCodeInput struct {
	Body struct {
		Code string `json:"code" required:"true" pattern:"^[0-9A-F]{8}-[0-9A-F]{8}$" minLength:"17" maxLength:"17" example:"A1B2C3D4-E5F6A7B8"`
	}
}

// Exactly one of Code (TOTP) or RecoveryCode must be provided alongside Password
type twoFADestructiveInputBody struct {
	Password     string  `json:"password" required:"true" minLength:"8" maxLength:"72"`
	Code         *string `json:"code,omitempty" pattern:"^\\d{6}$" minLength:"6" maxLength:"6" example:"123456"`
	RecoveryCode *string `json:"recovery_code,omitempty" pattern:"^[0-9A-F]{8}-[0-9A-F]{8}$" minLength:"17" maxLength:"17" example:"A1B2C3D4-E5F6A7B8"`
}

type twoFADestructiveInput struct {
	Body twoFADestructiveInputBody
}

// ===== Handler =====

type TwoFactor struct {
	svc *service.TwoFactor
}

func NewTwoFactor(svc *service.TwoFactor) *TwoFactor {
	return &TwoFactor{svc: svc}
}

func (h *TwoFactor) Register(api huma.API) {
	huma.Register(api, huma.Operation{
		OperationID: "setup-2fa",
		Method:      http.MethodPost,
		Path:        "/auth/2fa/setup",
		Summary:     "2FA Setup",
		Description: "The first step to initialize 2FA for the current account.",
		Tags:        []string{"Auth"},
		Middlewares: huma.Middlewares{middleware.RateLimit(api, middleware.HeavyAuthLimiter)},
	}, h.setup)

	huma.Register(api, huma.Operation{
		OperationID: "confirm-2fa",
		Method:      http.MethodPost,
		Path:        "/auth/2fa/confirm",
		Summary:     "2FA Confirmation",
		Description: "Confirms that the user has the correct secret to generate the correct codes.",
		Tags:        []string{"Auth"},
		Middlewares: huma.Middlewares{middleware.RateLimit(api, middleware.HeavyAuthLimiter)},
	}, h.confirm)

	huma.Register(api, huma.Operation{
		OperationID: "verify-2fa",
		Method:      http.MethodPost,
		Path:        "/auth/2fa/verify",
		Summary:     "2FA Verification",
		Description: "The normal endpoint for a normal login, with 2fa already set up.",
		Tags:        []string{"Auth"},
		Middlewares: huma.Middlewares{middleware.RateLimit(api, middleware.HeavyAuthLimiter)},
	}, h.verify)

	huma.Register(api, huma.Operation{
		OperationID: "verify-2fa-recovery",
		Method:      http.MethodPost,
		Path:        "/auth/2fa/verify-recovery",
		Summary:     "2FA Recovery Code Verification",
		Description: "Use a one-time recovery code instead of a TOTP code to complete login.",
		Tags:        []string{"Auth"},
		Middlewares: huma.Middlewares{middleware.RateLimit(api, middleware.HeavyAuthLimiter)},
	}, h.verifyRecovery)

	huma.Register(api, huma.Operation{
		OperationID: "regenerate-2fa-codes",
		Method:      http.MethodPost,
		Path:        "/auth/2fa/regenerate-codes",
		Summary:     "Regenerate 2FA Recovery Codes",
		Description: "Invalidate all existing recovery codes and issue a fresh set. Requires the current password and either a TOTP code or a recovery code.",
		Tags:        []string{"Auth"},
		Middlewares: huma.Middlewares{middleware.RateLimit(api, middleware.HeavyAuthLimiter)},
	}, h.regenerateCodes)

	huma.Register(api, huma.Operation{
		OperationID: "disable-2fa",
		Method:      http.MethodDelete,
		Path:        "/auth/2fa",
		Summary:     "Disable 2FA",
		Description: "Remove 2FA from the account. Requires the current password and either a TOTP code or a recovery code.",
		Tags:        []string{"Auth"},
		Middlewares: huma.Middlewares{middleware.RateLimit(api, middleware.HeavyAuthLimiter)},
	}, h.disable)
}

func (h *TwoFactor) setup(ctx context.Context, _ *struct{}) (*twoFASetupOutput, error) {
	otpKey, err := h.svc.Setup(ctx)
	if err != nil {
		return nil, toHumaErr(err, "", "failed to run 2fa setup")
	}

	return &twoFASetupOutput{
		Body: twoFASetupOutputBody{
			Secret:     otpKey.Secret(),
			OTPAuthURL: otpKey.URL(),
		},
	}, nil
}

func (h *TwoFactor) confirm(ctx context.Context, in *twoFACodeInput) (*twoFAConfirmOutput, error) {
	codes, err := h.svc.Confirm(ctx, in.Body.Code)
	if err != nil {
		return nil, toHumaErr(err, "", "failed to confirm 2fa setup")
	}

	return &twoFAConfirmOutput{
		Body: twoFAConfirmOutputBody{
			RecoveryCodes: codes,
		},
	}, nil
}

func (h *TwoFactor) verify(ctx context.Context, in *twoFACodeInput) (*struct{}, error) {
	err := h.svc.Verify(ctx, in.Body.Code)
	if err != nil {
		return nil, toHumaErr(err, "", "failed to verify 2fa")
	}

	return &struct{}{}, nil
}

func (h *TwoFactor) verifyRecovery(ctx context.Context, in *twoFARecoveryCodeInput) (*struct{}, error) {
	err := h.svc.VerifyRecoveryCode(ctx, in.Body.Code)
	if err != nil {
		return nil, toHumaErr(err, "", "failed to verify recovery code")
	}

	return &struct{}{}, nil
}

func (h *TwoFactor) regenerateCodes(ctx context.Context, in *twoFADestructiveInput) (*twoFAConfirmOutput, error) {
	codes, err := h.svc.RegenerateCodes(ctx, in.Body.Password, in.Body.Code, in.Body.RecoveryCode)
	if err != nil {
		return nil, toHumaErr(err, "", "failed to regenerate recovery codes")
	}

	return &twoFAConfirmOutput{
		Body: twoFAConfirmOutputBody{
			RecoveryCodes: codes,
		},
	}, nil
}

func (h *TwoFactor) disable(ctx context.Context, in *twoFADestructiveInput) (*struct{}, error) {
	err := h.svc.Disable(ctx, in.Body.Password, in.Body.Code, in.Body.RecoveryCode)
	if err != nil {
		return nil, toHumaErr(err, "", "failed to disable 2fa")
	}

	return &struct{}{}, nil
}
