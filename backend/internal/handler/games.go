package handler

import (
	"context"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/officeryoda/dozingo/internal/generated"
)

/// ===== Input/Output types =====

type GameOutput struct {
	ID       string  `json:"id" format:"uuid"`
	BoardID  *string `json:"board_id" format:"uuid"`
	PlayerID string  `json:"player_id" format:"uuid"`
	Status   string  `json:"status" format:"text"`
}

type GetGameByIDInput struct {
	ID string `path:"game_id" format:"uuid"`
}

type GetGameByIDOutput struct {
	Body GameOutput
}

type ListGamesByPlayerInput struct {
	PlayerID string `path:"player_id" format:"uuid"`
}

type ListGamesByPlayerOutput struct {
	Body []GameOutput
}

type ListGamesByBoardInput struct {
	BoardID string `path:"board_id" format:"uuid"`
}

type ListGamesByBoardOutput struct {
	Body []GameOutput
}

type CreateGameInput struct {
	PlayerID string `query:"player_id" format:"uuid"` // TODO eventually replace this when user auth is working
	BoardID  string `path:"board_id" format:"uuid"`
}

type CreateGameOutput struct {
	Body GameOutput
}

type UpdateGameStatusInput struct {
	GameID   string `path:"game_id" format:"uuid"`
	PlayerID string `query:"player_id" format:"uuid"` // TODO eventually replace this when user auth is working
	Body     struct {
		Status string `json:"status" format:"text" maxLength:"20" doc:"must be any of those: 'active', 'completed' or 'abandoned'"`
	}
}

type UpdateGameStatusOutput struct {
	Body GameOutput
}

type DeleteGameInput struct {
	GameID string `path:"game_id" format:"uuid"`
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
	id, err := uuidFromString(input.ID)
	if err != nil {
		return nil, huma.Error400BadRequest("invalid id", err)
	}

	cells, err := queries.GetGameByID(ctx, id)
	if err != nil {
		return nil, huma.Error500InternalServerError("internal server error", err)
	}

	return &GetGameByIDOutput{Body: gameToOutput(cells)}, nil
}

func listGamesByPlayer(ctx context.Context, queries *generated.Queries, input ListGamesByPlayerInput) (*ListGamesByPlayerOutput, error) {
	playerID, err := uuidFromString(input.PlayerID)
	if err != nil {
		return nil, huma.Error400BadRequest("invalid input_id", err)
	}

	games, err := queries.ListGamesByPlayer(ctx, playerID)
	if err != nil {
		return nil, huma.Error500InternalServerError("internal server error", err)
	}

	out := &ListGamesByPlayerOutput{}
	out.Body = make([]GameOutput, 0)
	for i := range games {
		out.Body = append(out.Body, gameToOutput(games[i]))
	}

	return out, nil
}

func listGamesByBoard(ctx context.Context, queries *generated.Queries, input ListGamesByBoardInput) (*ListGamesByBoardOutput, error) {
	boardID, err := uuidFromString(input.BoardID)
	if err != nil {
		return nil, huma.Error400BadRequest("invalid board_id", err)
	}

	games, err := queries.ListGamesByBoard(ctx, boardID)
	if err != nil {
		return nil, huma.Error500InternalServerError("internal server error", err)
	}

	out := &ListGamesByBoardOutput{}
	out.Body = make([]GameOutput, 0)
	for i := range games {
		out.Body = append(out.Body, gameToOutput(games[i]))
	}

	return out, nil
}

func createGame(ctx context.Context, queries *generated.Queries, input CreateGameInput) (*CreateGameOutput, error) {
	playerID, err := uuidFromString(input.PlayerID)
	if err != nil {
		return nil, huma.Error400BadRequest("invalid player_id", err)
	}

	boardID, err := uuidFromString(input.BoardID)
	if err != nil {
		return nil, huma.Error400BadRequest("invalid board_id", err)
	}

	game, err := queries.CreateGame(ctx, generated.CreateGameParams{
		PlayerID: playerID,
		BoardID:  boardID,
	})
	if err != nil {
		return nil, huma.Error500InternalServerError("failed to create game", err)
	}

	return &CreateGameOutput{Body: gameToOutput(game)}, nil
}

func updateGame(ctx context.Context, queries *generated.Queries, input UpdateGameStatusInput) (*UpdateGameStatusOutput, error) {
	gameID, err := uuidFromString(input.GameID)
	if err != nil {
		return nil, huma.Error400BadRequest("invalid game_id", err)
	}
	playerID, err := uuidFromString(input.PlayerID)
	if err != nil {
		return nil, huma.Error400BadRequest("invalid player_id", err)
	}

	game, err := queries.UpdateGameStatus(ctx, generated.UpdateGameStatusParams{
		ID:       gameID,
		PlayerID: playerID,
		Status:   input.Body.Status,
	})
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, huma.Error404NotFound("game not found", err)
		}
		return nil, huma.Error500InternalServerError("failed to update game", err)
	}

	return &UpdateGameStatusOutput{Body: gameToOutput(game)}, nil
}

func deleteGame(ctx context.Context, queries *generated.Queries, input DeleteGameInput) (*struct{}, error) {
	gameID, err := uuidFromString(input.GameID)
	if err != nil {
		return nil, huma.Error400BadRequest("invalid game_id", err)
	}

	_, err = queries.DeleteGame(ctx, gameID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, huma.Error404NotFound("game not found", err)
		}
		return nil, huma.Error500InternalServerError("failed to delete game", err)
	}

	return &struct{}{}, nil
}

/// ===== Helper =====

func gameToOutput(game generated.Game) GameOutput {
	return GameOutput{
		ID:       game.ID.String(),
		BoardID:  stringFromPgUUID(game.BoardID),
		PlayerID: game.PlayerID.String(),
		Status:   game.Status,
	}
}
