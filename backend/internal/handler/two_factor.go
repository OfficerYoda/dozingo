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

func (h *TwoFactor) confirm(ctx context.Context, in *twoFACodeInput) (*struct{}, error) {
	err := h.svc.Confirm(ctx, in.Body.Code)
	if err != nil {
		return nil, toHumaErr(err, "", "failed to confirm 2fa setup")
	}

	return &struct{}{}, nil
}

func (h *TwoFactor) verify(ctx context.Context, in *twoFACodeInput) (*struct{}, error) {
	err := h.svc.Verify(ctx, in.Body.Code)
	if err != nil {
		return nil, toHumaErr(err, "", "failed to verify 2fa")
	}

	return &struct{}{}, nil
}
