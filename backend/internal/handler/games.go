package handler

import (
	"context"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/officeryoda/dozingo/internal/generated"
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
	PlayerID types.UUIDParam `query:"player_id" format:"uuid"` // TODO eventually replace this when user auth is working
	BoardID  types.UUIDParam `path:"board_id" format:"uuid"`
}

type CreateGameOutput struct {
	Body GameOutput
}

type UpdateGameStatusInput struct {
	GameID   types.UUIDParam `path:"game_id" format:"uuid"`
	PlayerID types.UUIDParam `query:"player_id" format:"uuid"` // TODO eventually replace this when user auth is working
	Body     struct {
		Status string `json:"status" format:"text" maxLength:"20" doc:"must be any of those: 'active', 'completed' or 'abandoned'"`
	}
}

type UpdateGameStatusOutput struct {
	Body GameOutput
}

type DeleteGameInput struct {
	GameID types.UUIDParam `path:"game_id" format:"uuid"`
}

/// ===== Register =====

func RegisterGames(api huma.API, pool *pgxpool.Pool) {
	queries := generated.New(pool)

	huma.Register(api, huma.Operation{
		OperationID: "get-game-by-id",
		Method:      http.MethodGet,
		Path:        "/games/{game_id}",
		Summary:     "Get a Game by ID",
		Tags:        []string{"Games"},
	}, func(ctx context.Context, input *GetGameByIDInput) (*GetGameByIDOutput, error) {
		return getGameByID(ctx, queries, *input)
	})

	huma.Register(api, huma.Operation{
		OperationID: "list-games-by-player",
		Method:      http.MethodGet,
		Path:        "/users/{player_id}/games",
		Summary:     "List all games by a player",
		Tags:        []string{"Games"},
	}, func(ctx context.Context, input *ListGamesByPlayerInput) (*ListGamesByPlayerOutput, error) {
		return listGamesByPlayer(ctx, queries, *input)
	})

	huma.Register(api, huma.Operation{
		OperationID: "list-games-by-board",
		Method:      http.MethodGet,
		Path:        "/boards/{board_id}/games",
		Summary:     "List all games played on a board",
		Tags:        []string{"Games"},
	}, func(ctx context.Context, input *ListGamesByBoardInput) (*ListGamesByBoardOutput, error) {
		return listGamesByBoard(ctx, queries, *input)
	})

	huma.Register(api, huma.Operation{
		OperationID: "create-game",
		Method:      http.MethodPost,
		Path:        "/boards/{board_id}/games",
		Summary:     "Create a game",
		Tags:        []string{"Games"},
	}, func(ctx context.Context, input *CreateGameInput) (*CreateGameOutput, error) {
		return createGame(ctx, queries, *input)
	})

	huma.Register(api, huma.Operation{
		OperationID: "update-game-status",
		Method:      http.MethodPut,
		Path:        "/games/{game_id}/status",
		Summary:     "Update game status",
		Tags:        []string{"Games"},
	}, func(ctx context.Context, input *UpdateGameStatusInput) (*UpdateGameStatusOutput, error) {
		return updateGame(ctx, queries, *input)
	})

	huma.Register(api, huma.Operation{
		OperationID: "delete-game",
		Method:      http.MethodDelete,
		Path:        "/games/{game_id}",
		Summary:     "Delete a game",
		Tags:        []string{"Games"},
	}, func(ctx context.Context, input *DeleteGameInput) (*struct{}, error) {
		return deleteGame(ctx, queries, *input)
	})
}

/// ===== Handlers =====

func getGameByID(ctx context.Context, queries *generated.Queries, input GetGameByIDInput) (*GetGameByIDOutput, error) {
	game, err := queries.GetGameByID(ctx, input.GameID.Value)
	if err != nil {
		return nil, notFoundOr500(err, "game not found", "failed to get game")
	}

	return &GetGameByIDOutput{Body: gameToOutput(game)}, nil
}

func listGamesByPlayer(ctx context.Context, queries *generated.Queries, input ListGamesByPlayerInput) (*ListGamesByPlayerOutput, error) {
	games, err := queries.ListGamesByPlayer(ctx, input.PlayerID.Value)
	if err != nil {
		return nil, internalError(err, "failed to list games by player")
	}

	return &ListGamesByPlayerOutput{Body: mapSlice(games, gameToOutput)}, nil
}

func listGamesByBoard(ctx context.Context, queries *generated.Queries, input ListGamesByBoardInput) (*ListGamesByBoardOutput, error) {
	games, err := queries.ListGamesByBoard(ctx, input.BoardID.Value)
	if err != nil {
		return nil, internalError(err, "failed to list games by board")
	}

	return &ListGamesByBoardOutput{Body: mapSlice(games, gameToOutput)}, nil
}

func createGame(ctx context.Context, queries *generated.Queries, input CreateGameInput) (*CreateGameOutput, error) {
	game, err := queries.CreateGame(ctx, generated.CreateGameParams{
		PlayerID: input.PlayerID.Value,
		BoardID:  input.BoardID.Value,
	})
	if err != nil {
		return nil, internalError(err, "failed to create game")
	}

	return &CreateGameOutput{Body: gameToOutput(game)}, nil
}

func updateGame(ctx context.Context, queries *generated.Queries, input UpdateGameStatusInput) (*UpdateGameStatusOutput, error) {
	game, err := queries.UpdateGameStatus(ctx, generated.UpdateGameStatusParams{
		ID:       input.GameID.Value,
		PlayerID: input.PlayerID.Value,
		Status:   input.Body.Status,
	})
	if err != nil {
		return nil, notFoundOr500(err, "game not found", "failed to update game")
	}

	return &UpdateGameStatusOutput{Body: gameToOutput(game)}, nil
}

func deleteGame(ctx context.Context, queries *generated.Queries, input DeleteGameInput) (*struct{}, error) {
	_, err := queries.DeleteGame(ctx, input.GameID.Value)
	if err != nil {
		return nil, notFoundOr500(err, "game not found", "failed to delete game")
	}

	return &struct{}{}, nil
}

/// ===== Helper =====

func gameToOutput(game generated.Game) GameOutput {
	return GameOutput{
		GameID:   game.ID.String(),
		BoardID:  stringFromPgUUID(game.BoardID),
		PlayerID: game.PlayerID.String(),
		Status:   game.Status,
	}
}
