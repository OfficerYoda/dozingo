package service

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/officeryoda/dozingo/internal/generated"
	"github.com/officeryoda/dozingo/internal/repository"
)

type GameCells interface {
	ListByGameID(ctx context.Context, gameID pgtype.UUID) ([]generated.GameCell, error)
	Create(ctx context.Context, in repository.CreateGameCellsInput) ([]generated.GameCell, error)
	UpdateMark(ctx context.Context, in repository.UpdateGameCellMarkInput) (generated.GameCell, error)
}

type gameCells struct {
	repo *repository.GameCells
}

func NewGameCells(repo *repository.GameCells) GameCells {
	return &gameCells{repo: repo}
}

func (s *gameCells) ListByGameID(ctx context.Context, gameID pgtype.UUID) ([]generated.GameCell, error) {
	return s.repo.ListByGameID(ctx, gameID)
}

func (s *gameCells) Create(ctx context.Context, in repository.CreateGameCellsInput) ([]generated.GameCell, error) {
	// TODO(authz): once handlers pass the session, verify the caller owns
	// the game before creating cells on it.
	return s.repo.Create(ctx, in)
}

func (s *gameCells) UpdateMark(ctx context.Context, in repository.UpdateGameCellMarkInput) (generated.GameCell, error) {
	// TODO(authz): verify the caller owns the game before marking cells.
	return s.repo.UpdateMark(ctx, in)
}
