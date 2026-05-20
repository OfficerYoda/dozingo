package handler

import (
	"context"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/officeryoda/dozingo/internal/generated"
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

/// ===== Register =====

func RegisterCells(api huma.API, pool *pgxpool.Pool) {
	queries := generated.New(pool)

	huma.Register(api, huma.Operation{
		OperationID: "get-cells-by-board-id",
		Method:      http.MethodGet,
		Path:        "/boards/{board_id}/cells",
		Summary:     "Get all cells for a Board",
		Tags:        []string{"Cells"},
	}, func(ctx context.Context, input *GetCellsByBoardIDInput) (*GetCellsByBoardIDOutput, error) {
		return getCellsByBoardID(ctx, queries, *input)
	})

	huma.Register(api, huma.Operation{
		OperationID: "create-cell",
		Method:      http.MethodPost,
		Path:        "/boards/{board_id}/cells",
		Summary:     "Create a cell",
		Tags:        []string{"Cells"},
	}, func(ctx context.Context, input *CreateCellInput) (*CreateCellOutput, error) {
		return createCell(ctx, queries, *input)
	})

	huma.Register(api, huma.Operation{
		OperationID: "update-cell",
		Method:      http.MethodPut,
		Path:        "/boards/{board_id}/cells/{cell_id}",
		Summary:     "Update a cell",
		Tags:        []string{"Cells"},
	}, func(ctx context.Context, input *UpdateCellInput) (*UpdateCellOutput, error) {
		return updateCell(ctx, queries, *input)
	})

	huma.Register(api, huma.Operation{
		OperationID: "delete-cell",
		Method:      http.MethodDelete,
		Path:        "/boards/{board_id}/cells/{cell_id}",
		Summary:     "Delete a cell",
		Tags:        []string{"Cells"},
	}, func(ctx context.Context, input *DeleteCellInput) (*struct{}, error) {
		return deleteCell(ctx, queries, *input)
	})
}

/// ===== Handlers =====

func getCellsByBoardID(ctx context.Context, queries *generated.Queries, input GetCellsByBoardIDInput) (*GetCellsByBoardIDOutput, error) {
	cells, err := queries.GetCellsByBoardID(ctx, input.BoardID.Value)
	if err != nil {
		return nil, internalError(err, "failed to get cells")
	}

	return &GetCellsByBoardIDOutput{Body: mapSlice(cells, cellToOutput)}, nil
}

func createCell(ctx context.Context, queries *generated.Queries, input CreateCellInput) (*CreateCellOutput, error) {
	var value int32 = 1
	if input.Body.Value != nil {
		value = *input.Body.Value
	}

	board, err := queries.CreateCell(ctx, generated.CreateCellParams{
		BoardID: input.BoardID.Value,
		Content: input.Body.Content,
		Value:   value,
	})
	if err != nil {
		return nil, internalError(err, "failed to create cell")
	}

	return &CreateCellOutput{Body: cellToOutput(board)}, nil
}

func updateCell(ctx context.Context, queries *generated.Queries, input UpdateCellInput) (*UpdateCellOutput, error) {
	var content pgtype.Text
	if input.Body.Content != nil {
		content = pgtype.Text{String: *input.Body.Content, Valid: true}
	}

	var value pgtype.Int4
	if input.Body.Value != nil {
		value = pgtype.Int4{Int32: *input.Body.Value, Valid: true}
	}

	cell, err := queries.UpdateCell(ctx, generated.UpdateCellParams{
		CellID:  input.CellID.Value,
		BoardID: input.BoardID.Value,
		Content: content,
		Value:   value,
	})
	if err != nil {
		return nil, notFoundOr500(err, "cell not found on this board", "failed to update cell")
	}

	return &UpdateCellOutput{Body: cellToOutput(cell)}, nil
}

func deleteCell(ctx context.Context, queries *generated.Queries, input DeleteCellInput) (*struct{}, error) {
	_, err := queries.DeleteCell(ctx, generated.DeleteCellParams{
		ID:      input.CellID.Value,
		BoardID: input.BoardID.Value,
	})
	if err != nil {
		return nil, notFoundOr500(err, "cell not found on this board", "failed to delete cell")
	}

	return &struct{}{}, nil
}

/// ===== Helper =====

func cellToOutput(cell generated.Cell) CellOutput {
	return CellOutput{
		CellID:   cell.ID.String(),
		BoardID:  cell.BoardID.String(),
		Content:  cell.Content,
		AuthorID: stringFromPgUUID(cell.AuthorID),
		Value:    cell.Value,
	}
}
