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

type GameCellOutput struct {
	GameCellID string  `json:"game_cell_id" format:"uuid"`
	GameID     string  `json:"game_id" format:"uuid"`
	CellID     *string `json:"cell_id" format:"uuid"`
	Content    string  `json:"content" format:"text"`
	Position   int32   `json:"position" format:"integer"`
	IsMarked   bool    `json:"is_marked"`
}

type GetGameCellsByGameIDInput struct {
	GameID types.UUIDParam `path:"game_id" format:"uuid"`
}

type GetGameCellsByGameIDOutput struct {
	Body []GameCellOutput
}

type CreateGameCellsInput struct {
	GameID types.UUIDParam `path:"game_id" format:"uuid"`
	Body   []struct {
		CellID   types.UUIDParam `json:"cell_id" format:"uuid"`
		Content  string          `json:"content" format:"text" required:"true" maxLength:"200"`
		Position int32           `json:"position" format:"integer" required:"true"`
	}
}

type CreateGameCellsOutput struct {
	Body []GameCellOutput
}

type UpdateGameCellMarkInput struct {
	GameID     types.UUIDParam `path:"game_id" format:"uuid"`
	GameCellID types.UUIDParam `path:"game_cell_id" format:"uuid"`
	Body       struct {
		IsMarked bool `json:"is_marked" required:"true"`
	}
}

type UpdateGameCellMarkOutput struct {
	Body GameCellOutput
}

/// ===== Handler =====

type GameCellsHandler struct {
	svc *service.GameCells
}

func NewGameCellsHandler(svc *service.GameCells) *GameCellsHandler {
	return &GameCellsHandler{svc: svc}
}

func (h *GameCellsHandler) Register(api huma.API) {
	huma.Register(api, huma.Operation{
		OperationID: "get-game-cells-by-game-id",
		Method:      http.MethodGet,
		Path:        "/games/{game_id}/cells",
		Summary:     "Get all cells for a game",
		Tags:        []string{"Game Cells"},
	}, h.list)

	huma.Register(api, huma.Operation{
		OperationID: "create-game-cells",
		Method:      http.MethodPost,
		Path:        "/games/{game_id}/cells",
		Summary:     "Bulk create game cells",
		Tags:        []string{"Game Cells"},
	}, h.create)

	huma.Register(api, huma.Operation{
		OperationID: "update-game-cell-mark",
		Method:      http.MethodPut,
		Path:        "/games/{game_id}/cells/{game_cell_id}",
		Summary:     "Update game cell mark",
		Tags:        []string{"Game Cells"},
	}, h.updateMark)
}

func (h *GameCellsHandler) list(ctx context.Context, in *GetGameCellsByGameIDInput) (*GetGameCellsByGameIDOutput, error) {
	cells, err := h.svc.ListByGameID(ctx, in.GameID.Value)
	if err != nil {
		return nil, toHumaErr(err, "", "failed to get game cells")
	}

	return &GetGameCellsByGameIDOutput{Body: mapSlice(cells, gameCellToOutput)}, nil
}

func (h *GameCellsHandler) create(ctx context.Context, in *CreateGameCellsInput) (*CreateGameCellsOutput, error) {
	items := make([]service.CreateGameCellItem, 0, len(in.Body))
	for _, c := range in.Body {
		items = append(items, service.CreateGameCellItem{
			CellID:   c.CellID.Value,
			Content:  c.Content,
			Position: c.Position,
		})
	}

	cells, err := h.svc.Create(ctx, service.CreateGameCellsInput{
		GameID: in.GameID.Value,
		Items:  items,
	})
	if err != nil {
		return nil, toHumaErr(err, "", "failed to create game cells")
	}

	return &CreateGameCellsOutput{Body: mapSlice(cells, gameCellToOutput)}, nil
}

func (h *GameCellsHandler) updateMark(ctx context.Context, in *UpdateGameCellMarkInput) (*UpdateGameCellMarkOutput, error) {
	cell, err := h.svc.UpdateMark(ctx, service.UpdateGameCellMarkInput{
		GameCellID: in.GameCellID.Value,
		GameID:     in.GameID.Value,
		IsMarked:   in.Body.IsMarked,
	})
	if err != nil {
		return nil, toHumaErr(err, "game cell not found", "failed to update game cell")
	}

	return &UpdateGameCellMarkOutput{Body: gameCellToOutput(cell)}, nil
}

func gameCellToOutput(cell generated.GameCell) GameCellOutput {
	return GameCellOutput{
		GameCellID: cell.ID.String(),
		GameID:     cell.GameID.String(),
		CellID:     pgmap.StringFromPgUUID(cell.CellID),
		Content:    cell.Content,
		Position:   cell.Position,
		IsMarked:   cell.IsMarked,
	}
}
