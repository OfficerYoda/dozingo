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

func (s *GameCells) ListByGameID(ctx context.Context, gameID pgtype.UUID) ([]generated.GameCell, error) {
	return s.gameCells.ListByGameID(ctx, gameID)
}

func (s *GameCells) Create(ctx context.Context, in repository.CreateGameCellsInput) ([]generated.GameCell, error) {
	_, err := checkIfCallerOwnsGame(ctx, s.games, s.queries, in.GameID)
	if err != nil {
		return []generated.GameCell{}, err
	}

	return s.gameCells.Create(ctx, in)
}

func (s *GameCells) UpdateMark(ctx context.Context, in repository.UpdateGameCellMarkInput) (generated.GameCell, error) {
	_, err := checkIfCallerOwnsGame(ctx, s.games, s.queries, in.GameID)
	if err != nil {
		return generated.GameCell{}, err
	}

	return s.gameCells.UpdateMark(ctx, in)
}
