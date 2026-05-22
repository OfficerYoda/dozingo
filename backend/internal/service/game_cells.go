package service

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/officeryoda/dozingo/internal/generated"
	"github.com/officeryoda/dozingo/internal/repository"
)

type GameCells struct {
	gameCells *repository.GameCells
}

func NewGameCells(gameCells *repository.GameCells) *GameCells {
	return &GameCells{gameCells: gameCells}
}

func (s *GameCells) ListByGameID(ctx context.Context, gameID pgtype.UUID) ([]generated.GameCell, error) {
	return s.gameCells.ListByGameID(ctx, gameID)
}

func (s *GameCells) Create(ctx context.Context, in repository.CreateGameCellsInput) ([]generated.GameCell, error) {
	// TODO(authz): once handlers pass the session, verify the caller owns
	// the game before creating cells on it.
	return s.gameCells.Create(ctx, in)
}

func (s *GameCells) UpdateMark(ctx context.Context, in repository.UpdateGameCellMarkInput) (generated.GameCell, error) {
	// TODO(authz): verify the caller owns the game before marking cells.
	return s.gameCells.UpdateMark(ctx, in)
}
