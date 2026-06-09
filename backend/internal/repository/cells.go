package repository

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/officeryoda/dozingo/internal/generated"
	"github.com/officeryoda/dozingo/internal/pgmap"
)

type Cells struct {
	queries *generated.Queries
}

type CreateCellInput struct {
	BoardID pgtype.UUID
	Content string
	Value   int32
}

type UpdateCellInput struct {
	CellID  pgtype.UUID
	BoardID pgtype.UUID
	Content *string
	Value   *int32
}

type DeleteCellInput struct {
	CellID  pgtype.UUID
	BoardID pgtype.UUID
}

func (r *Cells) ListByBoardID(ctx context.Context, boardID pgtype.UUID) ([]generated.Cell, error) {
	cells, err := r.queries.GetCellsByBoardID(ctx, boardID)
	if err != nil {
		return []generated.Cell{}, pgmap.TranslatePgErr(err)
	}

	return cells, nil
}

func (r *Cells) GetByID(ctx context.Context, cellID pgtype.UUID) (generated.Cell, error) {
	cell, err := r.queries.GetCellByID(ctx, cellID)
	if err != nil {
		return generated.Cell{}, pgmap.TranslatePgErr(err)
	}
	return cell, nil
}

func (r *Cells) Create(ctx context.Context, in CreateCellInput) (generated.Cell, error) {
	cell, err := r.queries.CreateCell(ctx, generated.CreateCellParams{
		BoardID: in.BoardID,
		Content: in.Content,
		Value:   in.Value,
	})
	if err != nil {
		return generated.Cell{}, pgmap.TranslatePgErr(err)
	}

	return cell, nil
}

func (r *Cells) Update(ctx context.Context, in UpdateCellInput) (generated.Cell, error) {
	cell, err := r.queries.UpdateCell(ctx, generated.UpdateCellParams{
		CellID:  in.CellID,
		BoardID: in.BoardID,
		Content: pgmap.PgTextFromString(in.Content),
		Value:   pgmap.PgInt4FromInt32(in.Value),
	})
	if err != nil {
		return generated.Cell{}, pgmap.TranslatePgErr(err)
	}

	return cell, nil
}

func (r *Cells) Delete(ctx context.Context, in DeleteCellInput) (generated.Cell, error) {
	cell, err := r.queries.DeleteCell(ctx, generated.DeleteCellParams{
		ID:      in.CellID,
		BoardID: in.BoardID,
	})
	if err != nil {
		return generated.Cell{}, pgmap.TranslatePgErr(err)
	}

	return cell, nil
}
