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
	boards    *repository.Boards
	queries   *generated.Queries
}

func NewGameCells(repos *repository.Repos, queries *generated.Queries) *GameCells {
	return &GameCells{
		gameCells: repos.GameCells,
		games:     repos.Games,
		boards:    repos.Boards,
		queries:   queries,
	}
}

type UpdateGameCellMarkInput struct {
	GameCellID pgtype.UUID
	GameID     pgtype.UUID
	IsMarked   bool
}

// UpdateMarkResult is the return value of UpdateMark. It includes the updated
// cell plus bingo information computed server-side.
type UpdateMarkResult struct {
	Cell       generated.GameCell
	BingoCount int32
	BingoDelta int32
}

func (s *GameCells) ListByGameID(ctx context.Context, gameID pgtype.UUID) ([]generated.GameCell, error) {
	// Verify the game exists before listing its cells so that callers receive
	// 404 for an unknown game_id rather than an empty list.
	if _, err := s.games.Get(ctx, gameID); err != nil {
		return nil, err
	}

	return s.gameCells.ListByGameID(ctx, gameID)
}

func (s *GameCells) UpdateMark(ctx context.Context, in UpdateGameCellMarkInput) (UpdateMarkResult, error) {
	_, err := checkIfCallerOwnsGame(ctx, s.games, s.queries, in.GameID)
	if err != nil {
		return UpdateMarkResult{}, err
	}

	err = s.assertGameCellOnGame(ctx, in.GameCellID, in.GameID)
	if err != nil {
		return UpdateMarkResult{}, err
	}

	cell, err := s.gameCells.UpdateMark(ctx, repository.UpdateGameCellMarkInput(in))
	if err != nil {
		return UpdateMarkResult{}, err
	}

	gameRow, err := s.games.Get(ctx, in.GameID)
	if err != nil {
		return UpdateMarkResult{}, err
	}
	oldBingoCount := gameRow.BingoCount

	newBingoCount, err := s.computeAndSetBingoCount(ctx, in.GameID, gameRow.BoardID)
	if err != nil {
		return UpdateMarkResult{}, err
	}

	return UpdateMarkResult{
		Cell:       cell,
		BingoCount: newBingoCount,
		BingoDelta: newBingoCount - oldBingoCount,
	}, nil
}

func (s *GameCells) computeAndSetBingoCount(ctx context.Context, gameID, boardID pgtype.UUID) (int32, error) {
	if !boardID.Valid {
		gameRow, err := s.games.Get(ctx, gameID)
		if err != nil {
			return 0, err
		}

		return gameRow.BingoCount, nil
	}

	board, err := s.boards.Get(ctx, boardID)
	if err != nil {
		return 0, fmt.Errorf("get board for bingo check: %w", err)
	}

	cells, err := s.gameCells.ListByGameID(ctx, gameID)
	if err != nil {
		return 0, fmt.Errorf("list game cells for bingo check: %w", err)
	}

	newCount := countCompleteBingos(cells, board.Size)

	updatedGame, err := s.games.SetBingoCount(ctx, gameID, newCount)
	if err != nil {
		return 0, fmt.Errorf("set bingo count: %w", err)
	}

	return updatedGame.BingoCount, nil
}

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
