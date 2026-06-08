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

func gameCellsToRepoItems(items []CreateGameCellItem) []repository.CreateGameCellItem {
	result := make([]repository.CreateGameCellItem, len(items))
	for i, item := range items {
		result[i] = repository.CreateGameCellItem(item)
	}
	return result
}
