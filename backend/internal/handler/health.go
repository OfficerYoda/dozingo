package handler

import (
	"context"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
)

// ===== Input/Output types =====

type healthOutputBody struct {
	Status string `json:"status"`
}

type healthOutput struct {
	Body healthOutputBody
}

// ===== Handler =====

type HealthHandler struct{}

func NewHealthHandler() *HealthHandler {
	return &HealthHandler{}
}

func (h *HealthHandler) Register(api huma.API) {
	huma.Register(api, huma.Operation{
		OperationID: "health-check",
		Method:      http.MethodGet,
		Path:        "/health",
		Summary:     "Health check",
	}, h.handleHealth)
}

func (h *HealthHandler) handleHealth(ctx context.Context, _ *struct{}) (*healthOutput, error) {
	return &healthOutput{Body: healthOutputBody{Status: "ok"}}, nil
}
