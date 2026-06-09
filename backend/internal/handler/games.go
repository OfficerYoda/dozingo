package handler

import (
	"context"
	"net/http"

	"github.com/danielgtaylor/huma/v2"

	"github.com/officeryoda/dozingo/internal/generated"
	"github.com/officeryoda/dozingo/internal/middleware"
	"github.com/officeryoda/dozingo/internal/pgmap"
	"github.com/officeryoda/dozingo/internal/service"
	"github.com/officeryoda/dozingo/internal/types"
)

// ===== Input/Output types =====

type gameOutput struct {
	GameID    string  `json:"game_id" format:"uuid"`
	BoardID   *string `json:"board_id" format:"uuid"`
	PlayerID  *string `json:"player_id" format:"uuid"`
	SessionID *string `json:"session_id" format:"uuid"`
	Status    string  `json:"status"`
}

type getGameByIDInput struct {
	GameID types.UUIDParam `path:"game_id" format:"uuid"`
}

type getGameByIDOutput struct {
	Body gameOutput
}

type listGamesByPlayerInput struct {
	PlayerID types.UUIDParam `path:"player_id" format:"uuid"`
}

type listGamesOutput struct {
	Body []gameOutput
}

type listGamesByBoardInput struct {
	BoardID types.UUIDParam `path:"board_id" format:"uuid"`
}

type listGamesByBoardOutput struct {
	Body []gameOutput
}

type createGameInput struct {
	BoardID types.UUIDParam `path:"board_id" format:"uuid"`
}

type createGameOutput struct {
	Body gameOutput
}

type updateGameStatusInputBody struct {
	Status string `json:"status" enum:"active,completed,abandoned" doc:"Game lifecycle state. Must be one of: active, completed, abandoned."`
}

type updateGameStatusInput struct {
	GameID types.UUIDParam `path:"game_id" format:"uuid"`
	Body   updateGameStatusInputBody
}

type updateGameStatusOutput struct {
	Body gameOutput
}

type deleteGameInput struct {
	GameID types.UUIDParam `path:"game_id" format:"uuid"`
}

// ===== Handler =====

type GamesHandler struct {
	svc *service.Games
}

func NewGamesHandler(svc *service.Games) *GamesHandler {
	return &GamesHandler{svc: svc}
}

func (h *GamesHandler) Register(api huma.API) {
	huma.Register(api, huma.Operation{
		OperationID: "get-game-by-id",
		Method:      http.MethodGet,
		Path:        "/games/{game_id}",
		Summary:     "Get a Game by ID",
		Tags:        []string{"Games"},
		Middlewares: huma.Middlewares{middleware.RateLimit(api, middleware.ReadLimiter)},
	}, h.get)

	huma.Register(api, huma.Operation{
		OperationID: "list-games-by-player",
		Method:      http.MethodGet,
		Path:        "/users/{player_id}/games",
		Summary:     "List all games by a player",
		Tags:        []string{"Games"},
		Middlewares: huma.Middlewares{middleware.RateLimit(api, middleware.ReadLimiter)},
	}, h.listByPlayer)

	huma.Register(api, huma.Operation{
		OperationID: "list-games-by-current-session",
		Method:      http.MethodGet,
		Path:        "/me/games",
		Summary:     "List all games from current session",
		Tags:        []string{"Games"},
		Middlewares: huma.Middlewares{middleware.RateLimit(api, middleware.ReadLimiter)},
	}, h.listByCurrentSession)

	huma.Register(api, huma.Operation{
		OperationID: "list-games-by-board",
		Method:      http.MethodGet,
		Path:        "/boards/{board_id}/games",
		Summary:     "List all games played on a board",
		Tags:        []string{"Games"},
		Middlewares: huma.Middlewares{middleware.RateLimit(api, middleware.ReadLimiter)},
	}, h.listByBoard)

	huma.Register(api, huma.Operation{
		OperationID: "create-game",
		Method:      http.MethodPost,
		Path:        "/boards/{board_id}/games",
		Summary:     "Create a game",
		Tags:        []string{"Games"},
		Middlewares: huma.Middlewares{middleware.RateLimit(api, middleware.WriteLimiter)},
	}, h.create)

	huma.Register(api, huma.Operation{
		OperationID: "update-game-status",
		Method:      http.MethodPut,
		Path:        "/games/{game_id}/status",
		Summary:     "Update game status",
		Tags:        []string{"Games"},
		Middlewares: huma.Middlewares{middleware.RateLimit(api, middleware.WriteLimiter)},
	}, h.updateStatus)

	huma.Register(api, huma.Operation{
		OperationID: "delete-game",
		Method:      http.MethodDelete,
		Path:        "/games/{game_id}",
		Summary:     "Delete a game",
		Tags:        []string{"Games"},
		Middlewares: huma.Middlewares{middleware.RateLimit(api, middleware.WriteLimiter)},
	}, h.delete)
}

func (h *GamesHandler) get(ctx context.Context, in *getGameByIDInput) (*getGameByIDOutput, error) {
	game, err := h.svc.Get(ctx, in.GameID.Value)
	if err != nil {
		return nil, toHumaErr(err, "game not found", "failed to get game")
	}

	return &getGameByIDOutput{Body: gameToOutput(game)}, nil
}

func (h *GamesHandler) listByPlayer(ctx context.Context, in *listGamesByPlayerInput) (*listGamesOutput, error) {
	games, err := h.svc.ListByPlayer(ctx, in.PlayerID.Value)
	if err != nil {
		return nil, toHumaErr(err, "", "failed to list games by player")
	}

	return &listGamesOutput{Body: mapSlice(games, gameToOutput)}, nil
}

func (h *GamesHandler) listByCurrentSession(ctx context.Context, _ *struct{}) (*listGamesOutput, error) {
	games, err := h.svc.ListByCurrentSession(ctx)
	if err != nil {
		return nil, toHumaErr(err, "", "failed to list games by player")
	}

	return &listGamesOutput{Body: mapSlice(games, gameToOutput)}, nil
}

func (h *GamesHandler) listByBoard(ctx context.Context, in *listGamesByBoardInput) (*listGamesByBoardOutput, error) {
	games, err := h.svc.ListByBoard(ctx, in.BoardID.Value)
	if err != nil {
		return nil, toHumaErr(err, "", "failed to list games by board")
	}

	return &listGamesByBoardOutput{Body: mapSlice(games, gameToOutput)}, nil
}

func (h *GamesHandler) create(ctx context.Context, in *createGameInput) (*createGameOutput, error) {
	game, err := h.svc.Create(ctx, in.BoardID.Value)
	if err != nil {
		return nil, toHumaErr(err, "", "failed to create game")
	}

	return &createGameOutput{Body: gameToOutput(game)}, nil
}

func (h *GamesHandler) updateStatus(ctx context.Context, in *updateGameStatusInput) (*updateGameStatusOutput, error) {
	game, err := h.svc.UpdateStatus(ctx, service.UpdateGameStatusInput{
		GameID: in.GameID.Value,
		Status: in.Body.Status,
	})
	if err != nil {
		return nil, toHumaErr(err, "game not found", "failed to update game")
	}

	return &updateGameStatusOutput{Body: gameToOutput(game)}, nil
}

func (h *GamesHandler) delete(ctx context.Context, in *deleteGameInput) (*struct{}, error) {
	if err := h.svc.Delete(ctx, in.GameID.Value); err != nil {
		return nil, toHumaErr(err, "game not found", "failed to delete game")
	}

	return &struct{}{}, nil
}

func gameToOutput(game generated.Game) gameOutput {
	return gameOutput{
		GameID:    game.ID.String(),
		BoardID:   pgmap.StringFromPgUUID(game.BoardID),
		PlayerID:  pgmap.StringFromPgUUID(game.PlayerID),
		SessionID: pgmap.StringFromPgUUID(game.SessionID),
		Status:    string(game.Status),
	}
}
