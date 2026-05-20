package repository

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/officeryoda/dozingo/internal/generated"
)

type Cells struct {
	queries *generated.Queries
}

type CreateCellInput struct {
	BoardID pgtype.UUID
	Content string
	Value   *int32
}

type UpdateCellInput struct {
	CellID  pgtype.UUID
	BoardID pgtype.UUID
	Content *string
	Value   *int32
}

func (r *Cells) ListByBoardID(ctx context.Context, boardID pgtype.UUID) ([]generated.Cell, error) {
	cells, err := r.queries.GetCellsByBoardID(ctx, boardID)
	if err != nil {
		return []generated.Cell{}, translatePgErr(err)
	}
	return cells, nil
}

func (r *Cells) Create(ctx context.Context, in CreateCellInput) (generated.Cell, error) {
	var value int32
	if in.Value != nil {
		value = *in.Value
	}
	cell, err := r.queries.CreateCell(ctx, generated.CreateCellParams{
		BoardID: in.BoardID,
		Content: in.Content,
		Value:   value,
	})
	if err != nil {
		return generated.Cell{}, translatePgErr(err)
	}
	return cell, nil
}

func (r *Cells) Update(ctx context.Context, in UpdateCellInput) (generated.Cell, error) {
	cell, err := r.queries.UpdateCell(ctx, generated.UpdateCellParams{
		CellID:  in.CellID,
		BoardID: in.BoardID,
		Content: pgTextFromString(in.Content),
		Value:   pgInt4FromInt32(in.Value),
	})
	if err != nil {
		return generated.Cell{}, translatePgErr(err)
	}
	return cell, nil
}

func (r *Cells) Delete(ctx context.Context, cellID, boardID pgtype.UUID) (generated.Cell, error) {
	cell, err := r.queries.DeleteCell(ctx, generated.DeleteCellParams{
		ID:      cellID,
		BoardID: boardID,
	})
	if err != nil {
		return generated.Cell{}, translatePgErr(err)
	}
	return cell, nil
}
