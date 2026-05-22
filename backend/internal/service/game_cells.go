package service

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/officeryoda/dozingo/internal/generated"
	"github.com/officeryoda/dozingo/internal/repository"
)

type GameCells struct {
	gameCells *repository.GameCells
	games     *repository.Games
	queries   *generated.Queries
}

func NewGameCells(gameCells *repository.GameCells, games *repository.Games, queries *generated.Queries) *GameCells {
	return &GameCells{
		gameCells: gameCells,
		games:     games,
		queries:   queries,
	}
}

type CreateGameCellsInput struct {
	GameID pgtype.UUID
	Items  []CreateGameCellItem
}

type CreateGameCellItem struct {
	CellID   pgtype.UUID
	Content  string
	Position int32
}

type UpdateGameCellMarkInput struct {
	GameCellID pgtype.UUID
	GameID     pgtype.UUID
	IsMarked   bool
}

func (s *GameCells) ListByGameID(ctx context.Context, gameID pgtype.UUID) ([]generated.GameCell, error) {
	return s.gameCells.ListByGameID(ctx, gameID)
}

func (s *GameCells) Create(ctx context.Context, in CreateGameCellsInput) ([]generated.GameCell, error) {
	_, err := checkIfCallerOwnsGame(ctx, s.games, s.queries, in.GameID)
	if err != nil {
		return []generated.GameCell{}, err
	}

	return s.gameCells.Create(ctx, repository.CreateGameCellsInput{
		GameID: in.GameID,
		Items:  gameCellsToRepoItems(in.Items),
	})
}

func (s *GameCells) UpdateMark(ctx context.Context, in UpdateGameCellMarkInput) (generated.GameCell, error) {
	_, err := checkIfCallerOwnsGame(ctx, s.games, s.queries, in.GameID)
	if err != nil {
		return generated.GameCell{}, err
	}

	return s.gameCells.UpdateMark(ctx, repository.UpdateGameCellMarkInput(in))
}

func gameCellsToRepoItems(items []CreateGameCellItem) []repository.CreateGameCellItem {
	result := make([]repository.CreateGameCellItem, len(items))
	for i, item := range items {
		result[i] = repository.CreateGameCellItem(item)
	}
	return result
}
