package handler

import (
	"context"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/officeryoda/dozingo/internal/generated"
	"github.com/officeryoda/dozingo/internal/pgmap"
	"github.com/officeryoda/dozingo/internal/repository"
	"github.com/officeryoda/dozingo/internal/service"
	"github.com/officeryoda/dozingo/internal/types"
)

/// ===== Input/Output types =====

type CellOutput struct {
	CellID   string  `json:"cell_id" format:"uuid"`
	BoardID  string  `json:"board_id" format:"uuid"`
	Content  string  `json:"content" format:"text"`
	AuthorID *string `json:"author_id" format:"uuid"`
	Value    int32   `json:"value"`
}

type GetCellsByBoardIDInput struct {
	BoardID types.UUIDParam `path:"board_id"`
}

type GetCellsByBoardIDOutput struct {
	Body []CellOutput
}

type CreateCellInput struct {
	BoardID types.UUIDParam `path:"board_id" format:"uuid"`
	Body    struct {
		Content string `json:"content" format:"text" required:"true" maxLength:"200"`
		Value   *int32 `json:"value,omitempty"`
	}
}

type CreateCellOutput struct {
	Body CellOutput
}

type UpdateCellInput struct {
	BoardID types.UUIDParam `path:"board_id" format:"uuid"`
	CellID  types.UUIDParam `path:"cell_id" format:"uuid"`
	Body    struct {
		Content *string `json:"content,omitempty" maxLength:"200"`
		Value   *int32  `json:"value,omitempty"`
	}
}

type UpdateCellOutput struct {
	Body CellOutput
}

type DeleteCellInput struct {
	BoardID types.UUIDParam `path:"board_id" format:"uuid"`
	CellID  types.UUIDParam `path:"cell_id" format:"uuid"`
}

/// ===== Handler =====

type CellsHandler struct {
	svc service.Cells
}

func NewCellsHandler(svc service.Cells) *CellsHandler {
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

func (h *CellsHandler) list(ctx context.Context, in *GetCellsByBoardIDInput) (*GetCellsByBoardIDOutput, error) {
	cells, err := h.svc.ListByBoardID(ctx, in.BoardID.Value)
	if err != nil {
		return nil, toHumaErr(err, "", "failed to list cells")
	}
	return &GetCellsByBoardIDOutput{Body: mapSlice(cells, cellToOutput)}, nil
}

func (h *CellsHandler) create(ctx context.Context, in *CreateCellInput) (*CreateCellOutput, error) {
	cell, err := h.svc.Create(ctx, repository.CreateCellInput{
		BoardID: in.BoardID.Value,
		Content: in.Body.Content,
		Value:   in.Body.Value,
	})
	if err != nil {
		return nil, toHumaErr(err, "", "failed to create cell")
	}
	return &CreateCellOutput{Body: cellToOutput(cell)}, nil
}

func (h *CellsHandler) update(ctx context.Context, in *UpdateCellInput) (*UpdateCellOutput, error) {
	cell, err := h.svc.Update(ctx, repository.UpdateCellInput{
		CellID:  in.CellID.Value,
		BoardID: in.BoardID.Value,
		Content: in.Body.Content,
		Value:   in.Body.Value,
	})
	if err != nil {
		return nil, toHumaErr(err, "cell not found on this board", "failed to update cell")
	}
	return &UpdateCellOutput{Body: cellToOutput(cell)}, nil
}

func (h *CellsHandler) delete(ctx context.Context, in *DeleteCellInput) (*struct{}, error) {
	if err := h.svc.Delete(ctx, in.CellID.Value, in.BoardID.Value); err != nil {
		return nil, toHumaErr(err, "cell not found on this board", "failed to delete cell")
	}
	return &struct{}{}, nil
}

func cellToOutput(cell generated.Cell) CellOutput {
	return CellOutput{
		CellID:   cell.ID.String(),
		BoardID:  cell.BoardID.String(),
		Content:  cell.Content,
		AuthorID: pgmap.StringFromPgUUID(cell.AuthorID),
		Value:    cell.Value,
	}
}
