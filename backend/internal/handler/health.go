package handler

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/officeryoda/dozingo/internal/middleware"
)

const healthDBTimeout = 2 * time.Second

// ===== Input/Output types =====

type healthOutputBody struct {
	Status   string `json:"status"`
	Database string `json:"database"`
}

type healthOutput struct {
	Status int `json:"-"`
	Body   healthOutputBody
}

// ===== Handler =====

type HealthHandler struct {
	pool *pgxpool.Pool
}

func NewHealthHandler(pool *pgxpool.Pool) *HealthHandler {
	return &HealthHandler{pool: pool}
}

func (h *HealthHandler) Register(api huma.API) {
	huma.Register(api, huma.Operation{
		OperationID: "health-check",
		Method:      http.MethodGet,
		Path:        "/health",
		Summary:     "Health check",
		Middlewares: huma.Middlewares{middleware.RateLimit(api, middleware.HealthLimiter)},
	}, h.handleHealth)
}

func (h *HealthHandler) handleHealth(ctx context.Context, _ *struct{}) (*healthOutput, error) {
	pingCtx, cancel := context.WithTimeout(ctx, healthDBTimeout)
	defer cancel()

	if err := h.pool.Ping(pingCtx); err != nil {
		slog.Warn("health: db ping failed", "error", err)
		return &healthOutput{
			Status: http.StatusServiceUnavailable,
			Body: healthOutputBody{
				Status:   "degraded",
				Database: "down",
			},
		}, nil
	}

	return &healthOutput{
		Status: http.StatusOK,
		Body: healthOutputBody{
			Status:   "ok",
			Database: "ok",
		},
	}, nil
}
