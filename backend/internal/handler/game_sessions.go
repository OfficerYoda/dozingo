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

type heartbeatInput struct {
	GameID types.UUIDParam `path:"game_id" format:"uuid"`
}

type playtimeOutputBody struct {
	TotalSeconds int64 `json:"total_seconds"`
}

type playtimeOutput struct {
	Body playtimeOutputBody
}

type playtimeByGameInput struct {
	GameID types.UUIDParam `path:"game_id" format:"uuid"`
}

type playtimeByBoardInput struct {
	BoardID types.UUIDParam `path:"board_id" format:"uuid"`
}

type playtimeByPlayerInput struct {
	PlayerID types.UUIDParam `path:"player_id" format:"uuid"`
}

type totalPlaytimeInput struct {
	Duration types.DurationParam `query:"duration" doc:"Stats window e.g. 24h, 1.5h"`
}

// ===== Handler =====

type GameSessionsHandler struct {
	svc *service.GameSessions
}

func NewGameSessionsHandler(svc *service.GameSessions) *GameSessionsHandler {
	return &GameSessionsHandler{svc: svc}
}

func (h *GameSessionsHandler) Register(api huma.API) {
	huma.Register(api, huma.Operation{
		OperationID: "game-heartbeat",
		Method:      http.MethodPost,
		Path:        "/games/{game_id}/heartbeat",
		Summary:     "Heartbeat to keep a game's session alive",
		Tags:        []string{"Games"},
		Middlewares: huma.Middlewares{middleware.RateLimit(api, middleware.WriteLimiter)},
	}, h.heartbeat)

	huma.Register(api, huma.Operation{
		OperationID: "get-playtime-by-game",
		Method:      http.MethodGet,
		Path:        "/stats/playtime/games/{game_id}",
		Summary:     "Get total playtime for a game",
		Tags:        []string{"Stats"},
		Middlewares: huma.Middlewares{middleware.RateLimit(api, middleware.ReadLimiter)},
	}, h.playtimeByGame)

	huma.Register(api, huma.Operation{
		OperationID: "get-playtime-by-board",
		Method:      http.MethodGet,
		Path:        "/stats/playtime/boards/{board_id}",
		Summary:     "Get total playtime on a board",
		Description: "Aggregates playtime across all players.",
		Tags:        []string{"Stats"},
		Middlewares: huma.Middlewares{middleware.RateLimit(api, middleware.ReadLimiter)},
	}, h.playtimeByBoard)

	huma.Register(api, huma.Operation{
		OperationID: "get-playtime-by-player",
		Method:      http.MethodGet,
		Path:        "/stats/playtime/users/{player_id}",
		Summary:     "Get total playtime for a player",
		Tags:        []string{"Stats"},
		Middlewares: huma.Middlewares{middleware.RateLimit(api, middleware.ReadLimiter)},
	}, h.playtimeByPlayer)

	huma.Register(api, huma.Operation{
		OperationID: "get-playtime-me",
		Method:      http.MethodGet,
		Path:        "/stats/playtime/me",
		Summary:     "Get total playtime for the current user",
		Tags:        []string{"Stats"},
		Middlewares: huma.Middlewares{middleware.RateLimit(api, middleware.ReadLimiter)},
	}, h.playtimeMe)

	huma.Register(api, huma.Operation{
		OperationID: "get-total-playtime",
		Method:      http.MethodGet,
		Path:        "/stats/playtime",
		Summary:     "Get total playtime across all games within a time window",
		Description: "Aggregates playtime across all players.",
		Tags:        []string{"Stats"},
		Middlewares: huma.Middlewares{middleware.RateLimit(api, middleware.ReadListLimiter)},
	}, h.totalPlaytime)
}

func (h *GameSessionsHandler) heartbeat(ctx context.Context, in *heartbeatInput) (*struct{}, error) {
	if _, err := h.svc.Heartbeat(ctx, in.GameID.Value); err != nil {
		return nil, toHumaErr(err, "game not found", "failed to heartbeat game session")
	}

	return &struct{}{}, nil
}

func (h *GameSessionsHandler) playtimeByGame(ctx context.Context, in *playtimeByGameInput) (*playtimeOutput, error) {
	seconds, err := h.svc.GetPlaytimeByGame(ctx, in.GameID.Value)
	if err != nil {
		return nil, toHumaErr(err, "game not found", "failed to get playtime by game")
	}

	return &playtimeOutput{Body: playtimeOutputBody{TotalSeconds: seconds}}, nil
}

func (h *GameSessionsHandler) playtimeByBoard(ctx context.Context, in *playtimeByBoardInput) (*playtimeOutput, error) {
	seconds, err := h.svc.GetPlaytimeByBoard(ctx, in.BoardID.Value)
	if err != nil {
		return nil, toHumaErr(err, "", "failed to get playtime by board")
	}

	return &playtimeOutput{Body: playtimeOutputBody{TotalSeconds: seconds}}, nil
}

func (h *GameSessionsHandler) playtimeByPlayer(ctx context.Context, in *playtimeByPlayerInput) (*playtimeOutput, error) {
	seconds, err := h.svc.GetPlaytimeByPlayer(ctx, in.PlayerID.Value)
	if err != nil {
		return nil, toHumaErr(err, "", "failed to get playtime by player")
	}

	return &playtimeOutput{Body: playtimeOutputBody{TotalSeconds: seconds}}, nil
}

func (h *GameSessionsHandler) playtimeMe(ctx context.Context, _ *struct{}) (*playtimeOutput, error) {
	seconds, err := h.svc.GetPlaytimeForCurrentUser(ctx)
	if err != nil {
		return nil, toHumaErr(err, "", "failed to get playtime for current user")
	}

	return &playtimeOutput{Body: playtimeOutputBody{TotalSeconds: seconds}}, nil
}

func (h *GameSessionsHandler) totalPlaytime(ctx context.Context, in *totalPlaytimeInput) (*playtimeOutput, error) {
	seconds, err := h.svc.GetTotalPlaytime(ctx, in.Duration.Value)
	if err != nil {
		return nil, toHumaErr(err, "", "failed to get total playtime")
	}

	return &playtimeOutput{Body: playtimeOutputBody{TotalSeconds: seconds}}, nil
}
