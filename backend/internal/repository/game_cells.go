package repository

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/officeryoda/dozingo/internal/generated"
)

type GameCells struct {
	queries *generated.Queries
}

type CreateGameCellItem struct {
	CellID   pgtype.UUID
	Content  string
	Position int32
}

type CreateGameCellsInput struct {
	GameID pgtype.UUID
	Items  []CreateGameCellItem
}

type UpdateGameCellMarkInput struct {
	GameCellID pgtype.UUID
	GameID     pgtype.UUID
	IsMarked   bool
}

func (r *GameCells) ListByGameID(ctx context.Context, gameID pgtype.UUID) ([]generated.GameCell, error) {
	cells, err := r.queries.GetGameCellsByGameID(ctx, gameID)
	if err != nil {
		return []generated.GameCell{}, translatePgErr(err)
	}
	return cells, nil
}

func (r *GameCells) Create(ctx context.Context, in CreateGameCellsInput) ([]generated.GameCell, error) {
	params := make([]generated.CreateGameCellsParams, 0, len(in.Items))
	for _, item := range in.Items {
		params = append(params, generated.CreateGameCellsParams{
			GameID:   in.GameID,
			CellID:   item.CellID,
			Content:  item.Content,
			Position: item.Position,
		})
	}

	if _, err := r.queries.CreateGameCells(ctx, params); err != nil {
		return []generated.GameCell{}, translatePgErr(err)
	}

	cells, err := r.queries.GetGameCellsByGameID(ctx, in.GameID)
	if err != nil {
		return []generated.GameCell{}, translatePgErr(err)
	}
	return cells, nil
}

func (r *GameCells) UpdateMark(ctx context.Context, in UpdateGameCellMarkInput) (generated.GameCell, error) {
	cell, err := r.queries.UpdateGameCellMark(ctx, generated.UpdateGameCellMarkParams{
		IsMarked: in.IsMarked,
		ID:       in.GameCellID,
		GameID:   in.GameID,
	})
	if err != nil {
		return generated.GameCell{}, translatePgErr(err)
	}
	return cell, nil
}
