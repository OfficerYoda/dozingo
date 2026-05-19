package handler

import (
	"context"
	"errors"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/officeryoda/dozingo/internal/generated"
)

/// ===== Input/Output types =====

type GameCellOutput struct {
	ID       string  `json:"id" format:"uuid"`
	GameID   string  `json:"game_id" format:"uuid"`
	CellID   *string `json:"cell_id" format:"uuid"`
	Content  string  `json:"content" format:"text"`
	Position int32   `json:"position" format:"integer"`
	IsMarked bool    `json:"is_marked"`
}

type GetGameCellsByGameIDInput struct {
	GameID string `path:"game_id" format:"uuid"`
}

type GetGameCellsByGameIDOutput struct {
	Body []GameCellOutput
}

type CreateGameCellsInput struct {
	GameID string `path:"game_id" format:"uuid"`
	Body   []struct {
		CellID   string `json:"cell_id" format:"uuid"`
		Content  string `json:"content" format:"text" required:"true" maxLength:"200"`
		Position int32  `json:"position" format:"integer" required:"true"`
	}
}

type CreateGameCellsOutput struct {
	Body []GameCellOutput
}

type UpdateGameCellMarkInput struct {
	GameID     string `path:"game_id" format:"uuid"`
	GameCellID string `path:"game_cell_id" format:"uuid"`
	Body       struct {
		IsMarked bool `json:"is_marked" required:"true"`
	}
}

type UpdateGameCellMarkOutput struct {
	Body GameCellOutput
}

/// ===== Register =====

func RegisterGameCells(api huma.API, pool *pgxpool.Pool) {
	queries := generated.New(pool)

	huma.Register(api, huma.Operation{
		OperationID: "get-game-cells-by-game-id",
		Method:      http.MethodGet,
		Path:        "/games/{game_id}/cells",
		Summary:     "Get all cells for a game",
		Tags:        []string{"Game Cells"},
	}, func(ctx context.Context, input *GetGameCellsByGameIDInput) (*GetGameCellsByGameIDOutput, error) {
		return getGameCellsByGameID(ctx, queries, *input)
	})

	huma.Register(api, huma.Operation{
		OperationID: "create-game-cells",
		Method:      http.MethodPost,
		Path:        "/games/{game_id}/cells",
		Summary:     "Bulk create game cells",
		Tags:        []string{"Game Cells"},
	}, func(ctx context.Context, input *CreateGameCellsInput) (*CreateGameCellsOutput, error) {
		return createGameCells(ctx, queries, *input)
	})

	huma.Register(api, huma.Operation{
		OperationID: "update-game-cell-mark",
		Method:      http.MethodPut,
		Path:        "/games/{game_id}/cells/{game_cell_id}",
		Summary:     "Update game cell mark",
		Tags:        []string{"Game Cells"},
	}, func(ctx context.Context, input *UpdateGameCellMarkInput) (*UpdateGameCellMarkOutput, error) {
		return updateGameCellMark(ctx, queries, *input)
	})
}

/// ===== Handlers =====

func getGameCellsByGameID(ctx context.Context, queries *generated.Queries, input GetGameCellsByGameIDInput) (*GetGameCellsByGameIDOutput, error) {
	gameID, err := uuidFromString(input.GameID)
	if err != nil {
		return nil, huma.Error400BadRequest("invalid game_id", err)
	}

	cells, err := queries.GetGameCellsByGameID(ctx, gameID)
	if err != nil {
		return nil, huma.Error500InternalServerError("internal server error", err)
	}

	out := &GetGameCellsByGameIDOutput{}
	out.Body = make([]GameCellOutput, 0)
	for _, c := range cells {
		out.Body = append(out.Body, gameCellToOutput(c))
	}

	return out, nil
}

func createGameCells(ctx context.Context, queries *generated.Queries, input CreateGameCellsInput) (*CreateGameCellsOutput, error) {
	gameID, err := uuidFromString(input.GameID)
	if err != nil {
		return nil, huma.Error400BadRequest("invalid game_id", err)
	}

	params := make([]generated.CreateGameCellsParams, 0, len(input.Body))
	for _, c := range input.Body {
		cellID, err := uuidFromString(c.CellID)
		if err != nil {
			return nil, huma.Error400BadRequest("invalid cell_id", err)
		}
		params = append(params, generated.CreateGameCellsParams{
			GameID:   gameID,
			CellID:   cellID,
			Content:  c.Content,
			Position: c.Position,
		})
	}

	_, err = queries.CreateGameCells(ctx, params)
	if err != nil {
		return nil, huma.Error500InternalServerError("failed to create game cells", err)
	}

	// Fetch the newly created cells to return them
	cells, err := queries.GetGameCellsByGameID(ctx, gameID)
	if err != nil {
		return nil, huma.Error500InternalServerError("failed to fetch game cells", err)
	}

	out := &CreateGameCellsOutput{}
	out.Body = make([]GameCellOutput, 0)
	for _, c := range cells {
		out.Body = append(out.Body, gameCellToOutput(c))
	}

	return out, nil
}

func updateGameCellMark(ctx context.Context, queries *generated.Queries, input UpdateGameCellMarkInput) (*UpdateGameCellMarkOutput, error) {
	gameID, err := uuidFromString(input.GameID)
	if err != nil {
		return nil, huma.Error400BadRequest("invalid game_id", err)
	}
	gameCellID, err := uuidFromString(input.GameCellID)
	if err != nil {
		return nil, huma.Error400BadRequest("invalid game_cell_id", err)
	}

	cell, err := queries.UpdateGameCellMark(ctx, generated.UpdateGameCellMarkParams{
		IsMarked: input.Body.IsMarked,
		ID:       gameCellID,
		GameID:   gameID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, huma.Error404NotFound("game cell not found", err)
		}
		return nil, huma.Error500InternalServerError("failed to update game cell", err)
	}

	return &UpdateGameCellMarkOutput{Body: gameCellToOutput(cell)}, nil
}

/// ===== Helper =====

func gameCellToOutput(cell generated.GameCell) GameCellOutput {
	return GameCellOutput{
		ID:       cell.ID.String(),
		GameID:   cell.GameID.String(),
		CellID:   stringFromPgUUID(cell.CellID),
		Content:  cell.Content,
		Position: cell.Position,
		IsMarked: cell.IsMarked,
	}
}
