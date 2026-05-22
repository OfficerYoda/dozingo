package handler

import (
	"context"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/officeryoda/dozingo/internal/generated"
	"github.com/officeryoda/dozingo/internal/pgmap"
	"github.com/officeryoda/dozingo/internal/service"
	"github.com/officeryoda/dozingo/internal/types"
)

/// ===== Input/Output types =====

type GameOutput struct {
	GameID   string  `json:"game_id" format:"uuid"`
	BoardID  *string `json:"board_id" format:"uuid"`
	PlayerID string  `json:"player_id" format:"uuid"`
	Status   string  `json:"status" format:"text"`
}

type GetGameByIDInput struct {
	GameID types.UUIDParam `path:"game_id" format:"uuid"`
}

type GetGameByIDOutput struct {
	Body GameOutput
}

type ListGamesByPlayerInput struct {
	PlayerID types.UUIDParam `path:"player_id" format:"uuid"`
}

type ListGamesByPlayerOutput struct {
	Body []GameOutput
}

type ListGamesByBoardInput struct {
	BoardID types.UUIDParam `path:"board_id" format:"uuid"`
}

type ListGamesByBoardOutput struct {
	Body []GameOutput
}

type CreateGameInput struct {
	BoardID types.UUIDParam `path:"board_id" format:"uuid"`
}

type CreateGameOutput struct {
	Body GameOutput
}

type UpdateGameStatusInput struct {
	GameID types.UUIDParam `path:"game_id" format:"uuid"`
	Body   struct {
		Status string `json:"status" format:"text" maxLength:"20" doc:"must be any of those: 'active', 'completed' or 'abandoned'"`
	}
}

type UpdateGameStatusOutput struct {
	Body GameOutput
}

type DeleteGameInput struct {
	GameID types.UUIDParam `path:"game_id" format:"uuid"`
}

/// ===== Handler =====

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
	}, h.get)

	huma.Register(api, huma.Operation{
		OperationID: "list-games-by-player",
		Method:      http.MethodGet,
		Path:        "/users/{player_id}/games",
		Summary:     "List all games by a player",
		Tags:        []string{"Games"},
	}, h.listByPlayer)

	huma.Register(api, huma.Operation{
		OperationID: "list-games-by-board",
		Method:      http.MethodGet,
		Path:        "/boards/{board_id}/games",
		Summary:     "List all games played on a board",
		Tags:        []string{"Games"},
	}, h.listByBoard)

	huma.Register(api, huma.Operation{
		OperationID: "create-game",
		Method:      http.MethodPost,
		Path:        "/boards/{board_id}/games",
		Summary:     "Create a game",
		Tags:        []string{"Games"},
	}, h.create)

	huma.Register(api, huma.Operation{
		OperationID: "update-game-status",
		Method:      http.MethodPut,
		Path:        "/games/{game_id}/status",
		Summary:     "Update game status",
		Tags:        []string{"Games"},
	}, h.updateStatus)

	huma.Register(api, huma.Operation{
		OperationID: "delete-game",
		Method:      http.MethodDelete,
		Path:        "/games/{game_id}",
		Summary:     "Delete a game",
		Tags:        []string{"Games"},
	}, h.delete)
}

func (h *GamesHandler) get(ctx context.Context, in *GetGameByIDInput) (*GetGameByIDOutput, error) {
	game, err := h.svc.Get(ctx, in.GameID.Value)
	if err != nil {
		return nil, toHumaErr(err, "game not found", "failed to get game")
	}

	return &GetGameByIDOutput{Body: gameToOutput(game)}, nil
}

func (h *GamesHandler) listByPlayer(ctx context.Context, in *ListGamesByPlayerInput) (*ListGamesByPlayerOutput, error) {
	games, err := h.svc.ListByPlayer(ctx, in.PlayerID.Value)
	if err != nil {
		return nil, toHumaErr(err, "", "failed to list games by player")
	}

	return &ListGamesByPlayerOutput{Body: mapSlice(games, gameToOutput)}, nil
}

func (h *GamesHandler) listByBoard(ctx context.Context, in *ListGamesByBoardInput) (*ListGamesByBoardOutput, error) {
	games, err := h.svc.ListByBoard(ctx, in.BoardID.Value)
	if err != nil {
		return nil, toHumaErr(err, "", "failed to list games by board")
	}

	return &ListGamesByBoardOutput{Body: mapSlice(games, gameToOutput)}, nil
}

func (h *GamesHandler) create(ctx context.Context, in *CreateGameInput) (*CreateGameOutput, error) {
	game, err := h.svc.Create(ctx, in.BoardID.Value)
	if err != nil {
		return nil, toHumaErr(err, "", "failed to create game")
	}

	return &CreateGameOutput{Body: gameToOutput(game)}, nil
}

func (h *GamesHandler) updateStatus(ctx context.Context, in *UpdateGameStatusInput) (*UpdateGameStatusOutput, error) {
	game, err := h.svc.UpdateStatus(ctx, service.UpdateGameStatusInput{
		GameID: in.GameID.Value,
		Status: in.Body.Status,
	})
	if err != nil {
		return nil, toHumaErr(err, "game not found", "failed to update game")
	}

	return &UpdateGameStatusOutput{Body: gameToOutput(game)}, nil
}

func (h *GamesHandler) delete(ctx context.Context, in *DeleteGameInput) (*struct{}, error) {
	if err := h.svc.Delete(ctx, in.GameID.Value); err != nil {
		return nil, toHumaErr(err, "game not found", "failed to delete game")
	}

	return &struct{}{}, nil
}

func gameToOutput(game generated.Game) GameOutput {
	return GameOutput{
		GameID:   game.ID.String(),
		BoardID:  pgmap.StringFromPgUUID(game.BoardID),
		PlayerID: game.PlayerID.String(),
		Status:   game.Status,
	}
}
