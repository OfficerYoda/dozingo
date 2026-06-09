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

type gameCellOutput struct {
	GameCellID string  `json:"game_cell_id" format:"uuid"`
	GameID     string  `json:"game_id" format:"uuid"`
	CellID     *string `json:"cell_id" format:"uuid"`
	Content    string  `json:"content"`
	Position   int32   `json:"position"`
	IsMarked   bool    `json:"is_marked"`
}

type getGameCellsByGameIDInput struct {
	GameID types.UUIDParam `path:"game_id" format:"uuid"`
}

type getGameCellsByGameIDOutput struct {
	Body []gameCellOutput
}

type updateGameCellMarkInputBody struct {
	IsMarked bool `json:"is_marked" required:"true"`
}

type updateGameCellMarkInput struct {
	GameID     types.UUIDParam `path:"game_id" format:"uuid"`
	GameCellID types.UUIDParam `path:"game_cell_id" format:"uuid"`
	Body       updateGameCellMarkInputBody
}

type updateGameCellMarkOutput struct {
	Body gameCellOutput
}

// ===== Handler =====

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
		Middlewares: huma.Middlewares{middleware.RateLimit(api, middleware.ReadLimiter)},
	}, h.list)

	huma.Register(api, huma.Operation{
		OperationID: "update-game-cell-mark",
		Method:      http.MethodPut,
		Path:        "/games/{game_id}/cells/{game_cell_id}",
		Summary:     "Update game cell mark",
		Tags:        []string{"Game Cells"},
		Middlewares: huma.Middlewares{middleware.RateLimit(api, middleware.GameplayLimiter)},
	}, h.updateMark)
}

func (h *GameCellsHandler) list(ctx context.Context, in *getGameCellsByGameIDInput) (*getGameCellsByGameIDOutput, error) {
	cells, err := h.svc.ListByGameID(ctx, in.GameID.Value)
	if err != nil {
		return nil, toHumaErr(err, "", "failed to get game cells")
	}

	return &getGameCellsByGameIDOutput{Body: mapSlice(cells, gameCellToOutput)}, nil
}

func (h *GameCellsHandler) updateMark(ctx context.Context, in *updateGameCellMarkInput) (*updateGameCellMarkOutput, error) {
	cell, err := h.svc.UpdateMark(ctx, service.UpdateGameCellMarkInput{
		GameCellID: in.GameCellID.Value,
		GameID:     in.GameID.Value,
		IsMarked:   in.Body.IsMarked,
	})
	if err != nil {
		return nil, toHumaErr(err, "game cell not found", "failed to update game cell")
	}

	return &updateGameCellMarkOutput{Body: gameCellToOutput(cell)}, nil
}

func gameCellToOutput(cell generated.GameCell) gameCellOutput {
	return gameCellOutput{
		GameCellID: cell.ID.String(),
		GameID:     cell.GameID.String(),
		CellID:     pgmap.StringFromPgUUID(cell.CellID),
		Content:    cell.Content,
		Position:   cell.Position,
		IsMarked:   cell.IsMarked,
	}
}
