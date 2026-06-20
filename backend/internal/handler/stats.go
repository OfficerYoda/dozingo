package handler

import (
	"context"
	"net/http"

	"github.com/danielgtaylor/huma/v2"

	"github.com/officeryoda/dozingo/internal/middleware"
	"github.com/officeryoda/dozingo/internal/service"
	"github.com/officeryoda/dozingo/internal/types"
)

// ===== Input/Output types =====

type recentStatsOutputBody struct {
	Bingos int `json:"bingos"`
	Games  int `json:"games"`
	Boards int `json:"boards"`
	Cells  int `json:"cells"`
}

type recentStatsOutput struct {
	Body recentStatsOutputBody
}

type recentStatsInput struct {
	Duration types.DurationParam `query:"duration" doc:"Stats window e.g. 24h, 1.5h"`
}

// ===== Handler =====

type StatsHandler struct {
	svc *service.Stats
}

func NewStatsHandler(svc *service.Stats) *StatsHandler {
	return &StatsHandler{svc: svc}
}

func (h *StatsHandler) Register(api huma.API) {
	huma.Register(api, huma.Operation{
		OperationID: "get-recent-stats",
		Method:      http.MethodGet,
		Path:        "/stats/recent",
		Summary:     "Get recent stats",
		Tags:        []string{"Stats"},
		Middlewares: huma.Middlewares{middleware.RateLimit(api, middleware.ReadListLimiter)},
	}, h.recent)
}

func (h *StatsHandler) recent(ctx context.Context, in *recentStatsInput) (*recentStatsOutput, error) {
	stats, err := h.svc.GetRecentStats(ctx, in.Duration.Value)
	if err != nil {
		return nil, toHumaErr(err, "", "failed to get recent stats")
	}

	// stats.Bingos is interface{} because sqlc cannot infer the type of
	// COALESCE(SUM(...), 0). PostgreSQL returns it as int64.
	var bingos int64
	if v, ok := stats.Bingos.(int64); ok {
		bingos = v
	}

	return &recentStatsOutput{
		Body: recentStatsOutputBody{
			Bingos: int(bingos),
			Games:  int(stats.Games),
			Boards: int(stats.Boards),
			Cells:  int(stats.Cells),
		},
	}, nil
}
