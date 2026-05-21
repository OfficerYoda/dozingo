package repository

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/officeryoda/dozingo/internal/generated"
	"github.com/officeryoda/dozingo/internal/pgmap"
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
		return []generated.GameCell{}, pgmap.TranslatePgErr(err)
	}
	return cells, nil
}

func (r *GameCells) Create(ctx context.Context, in CreateGameCellsInput) ([]generated.GameCell, error) {
	gameIDs := make([]pgtype.UUID, 0, len(in.Items))
	cellIDs := make([]pgtype.UUID, 0, len(in.Items))
	contents := make([]string, 0, len(in.Items))
	positions := make([]int32, 0, len(in.Items))
	for _, item := range in.Items {
		gameIDs = append(gameIDs, in.GameID)
		cellIDs = append(cellIDs, item.CellID)
		contents = append(contents, item.Content)
		positions = append(positions, item.Position)
	}

	gameCells, err := r.queries.CreateGameCells(ctx, generated.CreateGameCellsParams{
		GameIds:   gameIDs,
		CellIds:   cellIDs,
		Contents:  contents,
		Positions: positions,
	})
	if err != nil {
		return []generated.GameCell{}, pgmap.TranslatePgErr(err)
	}

	return gameCells, nil
}

func (r *GameCells) UpdateMark(ctx context.Context, in UpdateGameCellMarkInput) (generated.GameCell, error) {
	cell, err := r.queries.UpdateGameCellMark(ctx, generated.UpdateGameCellMarkParams{
		IsMarked: in.IsMarked,
		ID:       in.GameCellID,
		GameID:   in.GameID,
	})
	if err != nil {
		return generated.GameCell{}, pgmap.TranslatePgErr(err)
	}
	return cell, nil
}
