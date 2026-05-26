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

// ===== Input/Output types =====

type cellOutput struct {
	CellID   string  `json:"cell_id" format:"uuid"`
	BoardID  string  `json:"board_id" format:"uuid"`
	Content  string  `json:"content"`
	AuthorID *string `json:"author_id" format:"uuid"`
	Value    int32   `json:"value"`
}

type getCellsByBoardIDInput struct {
	BoardID types.UUIDParam `path:"board_id" format:"uuid"`
}

type getCellsByBoardIDOutput struct {
	Body []cellOutput
}

type createCellInputBody struct {
	Content string `json:"content" required:"true" maxLength:"200"`
	Value   *int32 `json:"value,omitempty"`
}

type createCellInput struct {
	BoardID types.UUIDParam `path:"board_id" format:"uuid"`
	Body    createCellInputBody
}

type createCellOutput struct {
	Body cellOutput
}

type updateCellInputBody struct {
	Content *string `json:"content,omitempty" maxLength:"200"`
	Value   *int32  `json:"value,omitempty"`
}

type updateCellInput struct {
	BoardID types.UUIDParam `path:"board_id" format:"uuid"`
	CellID  types.UUIDParam `path:"cell_id" format:"uuid"`
	Body    updateCellInputBody
}

type updateCellOutput struct {
	Body cellOutput
}

type deleteCellInput struct {
	BoardID types.UUIDParam `path:"board_id" format:"uuid"`
	CellID  types.UUIDParam `path:"cell_id" format:"uuid"`
}

// ===== Handler =====

type CellsHandler struct {
	svc *service.Cells
}

func NewCellsHandler(svc *service.Cells) *CellsHandler {
	return &CellsHandler{svc: svc}
}

func (h *CellsHandler) Register(api huma.API) {
	huma.Register(api, huma.Operation{
		OperationID: "get-cells-by-board-id",
		Method:      http.MethodGet,
		Path:        "/boards/{board_id}/cells",
		Summary:     "Get all cells for a Board",
		Tags:        []string{"Cells"},
	}, h.list)

	huma.Register(api, huma.Operation{
		OperationID: "create-cell",
		Method:      http.MethodPost,
		Path:        "/boards/{board_id}/cells",
		Summary:     "Create a cell",
		Tags:        []string{"Cells"},
	}, h.create)

	huma.Register(api, huma.Operation{
		OperationID: "update-cell",
		Method:      http.MethodPut,
		Path:        "/boards/{board_id}/cells/{cell_id}",
		Summary:     "Update a cell",
		Tags:        []string{"Cells"},
	}, h.update)

	huma.Register(api, huma.Operation{
		OperationID: "delete-cell",
		Method:      http.MethodDelete,
		Path:        "/boards/{board_id}/cells/{cell_id}",
		Summary:     "Delete a cell",
		Tags:        []string{"Cells"},
	}, h.delete)
}

func (h *CellsHandler) list(ctx context.Context, in *getCellsByBoardIDInput) (*getCellsByBoardIDOutput, error) {
	cells, err := h.svc.ListByBoardID(ctx, in.BoardID.Value)
	if err != nil {
		return nil, toHumaErr(err, "", "failed to list cells")
	}

	return &getCellsByBoardIDOutput{Body: mapSlice(cells, cellToOutput)}, nil
}

func (h *CellsHandler) create(ctx context.Context, in *createCellInput) (*createCellOutput, error) {
	cell, err := h.svc.Create(ctx, service.CreateCellInput{
		BoardID: in.BoardID.Value,
		Content: in.Body.Content,
		Value:   in.Body.Value,
	})
	if err != nil {
		return nil, toHumaErr(err, "", "failed to create cell")
	}

	return &createCellOutput{Body: cellToOutput(cell)}, nil
}

func (h *CellsHandler) update(ctx context.Context, in *updateCellInput) (*updateCellOutput, error) {
	cell, err := h.svc.Update(ctx, service.UpdateCellInput{
		CellID:  in.CellID.Value,
		BoardID: in.BoardID.Value,
		Content: in.Body.Content,
		Value:   in.Body.Value,
	})
	if err != nil {
		return nil, toHumaErr(err, "cell not found on this board", "failed to update cell")
	}

	return &updateCellOutput{Body: cellToOutput(cell)}, nil
}

func (h *CellsHandler) delete(ctx context.Context, in *deleteCellInput) (*struct{}, error) {
	err := h.svc.Delete(ctx, service.DeleteCellInput{
		CellID:  in.CellID.Value,
		BoardID: in.BoardID.Value,
	})
	if err != nil {
		return nil, toHumaErr(err, "cell not found on this board", "failed to delete cell")
	}

	return &struct{}{}, nil
}

func cellToOutput(cell generated.Cell) cellOutput {
	return cellOutput{
		CellID:   cell.ID.String(),
		BoardID:  cell.BoardID.String(),
		Content:  cell.Content,
		AuthorID: pgmap.StringFromPgUUID(cell.AuthorID),
		Value:    cell.Value,
	}
}
