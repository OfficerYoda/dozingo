package service

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/officeryoda/dozingo/internal/domain"
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

type UpdateGameCellMarkInput struct {
	GameCellID pgtype.UUID
	GameID     pgtype.UUID
	IsMarked   bool
}

func (s *GameCells) ListByGameID(ctx context.Context, gameID pgtype.UUID) ([]generated.GameCell, error) {
	// Verify the game exists before listing its cells so that callers receive
	// 404 for an unknown game_id rather than an empty list.
	if _, err := s.games.Get(ctx, gameID); err != nil {
		return nil, err
	}
	return s.gameCells.ListByGameID(ctx, gameID)
}

func (s *GameCells) UpdateMark(ctx context.Context, in UpdateGameCellMarkInput) (generated.GameCell, error) {
	_, err := checkIfCallerOwnsGame(ctx, s.games, s.queries, in.GameID)
	if err != nil {
		return generated.GameCell{}, err
	}

	if err := s.assertGameCellOnGame(ctx, in.GameCellID, in.GameID); err != nil {
		return generated.GameCell{}, err
	}

	return s.gameCells.UpdateMark(ctx, repository.UpdateGameCellMarkInput(in))
}

// assertGameCellOnGame fetches the game_cell by its id and verifies it
// really belongs to the game the caller named in the URL. The
// repository's UpdateGameCellMark SQL already scopes to (id, game_id)
// in the WHERE clause and would reject a cross-game id with no-rows.
// This extra check is defense-in-depth: it surfaces the error as an
// explicit "cell not found in this game" before the mutation runs and
// protects against a future SQL regression that drops the game_id
// scoping.
func (s *GameCells) assertGameCellOnGame(ctx context.Context, gameCellID, gameID pgtype.UUID) error {
	cell, err := s.gameCells.GetByID(ctx, gameCellID)
	if err != nil {
		return err
	}
	if cell.GameID != gameID {
		return fmt.Errorf("game cell does not belong to game: %w", domain.ErrNotFound)
	}

	return nil
}
