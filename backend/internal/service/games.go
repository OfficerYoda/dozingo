package service

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/officeryoda/dozingo/internal/domain"
	"github.com/officeryoda/dozingo/internal/generated"
	"github.com/officeryoda/dozingo/internal/middleware"
	"github.com/officeryoda/dozingo/internal/repository"
)

type Games struct {
	games     *repository.Games
	gameCells *repository.GameCells
	boards    *repository.Boards
	cells     *repository.Cells
	queries   *generated.Queries
	txRunner  repository.TxRunner
}

func NewGames(
	repos *repository.Repos,
	queries *generated.Queries,
	txRunner repository.TxRunner,
) *Games {
	return &Games{
		games:     repos.Games,
		gameCells: repos.GameCells,
		boards:    repos.Boards,
		cells:     repos.Cells,
		queries:   queries,
		txRunner:  txRunner,
	}
}

type CreateGameInput struct {
	BoardID   pgtype.UUID
	GameCells []CreateGameCellItem
}

type CreateGameCellItem struct {
	CellID   pgtype.UUID
	Position int32
}

type UpdateGameStatusInput struct {
	GameID pgtype.UUID
	Status string
}

func (s *Games) Get(ctx context.Context, gameID pgtype.UUID) (generated.Game, error) {
	return s.games.Get(ctx, gameID)
}

func (s *Games) ListBySession(ctx context.Context) ([]generated.Game, error) {
	sessionUser, err := requiresVerifiedSession(ctx, s.queries)
	if err != nil {
		return []generated.Game{}, err
	}

	return s.ListByPlayer(ctx, sessionUser.UserID)
}

func (s *Games) ListByPlayer(ctx context.Context, playerID pgtype.UUID) ([]generated.Game, error) {
	return s.games.ListByPlayer(ctx, playerID)
}

func (s *Games) ListByCurrentSession(ctx context.Context) ([]generated.Game, error) {
	sessionUser, err := middleware.RequireSession(ctx, s.queries)
	if err != nil {
		return []generated.Game{}, err
	}

	return s.games.ListBySession(ctx, sessionUser.SessionID)
}

func (s *Games) ListByBoard(ctx context.Context, boardID pgtype.UUID) ([]generated.Game, error) {
	return s.games.ListByBoard(ctx, boardID)
}

func (s *Games) Create(ctx context.Context, in CreateGameInput) (generated.Game, error) {
	sessionUser, err := middleware.RequireSession(ctx, s.queries)
	if err != nil {
		return generated.Game{}, err
	}

	board, err := s.boards.Get(ctx, in.BoardID)
	if err != nil {
		return generated.Game{}, fmt.Errorf("board does not exist: %w", domain.ErrBadInput)
	}

	if int(board.Size*board.Size) != len(in.GameCells) {
		return generated.Game{}, fmt.Errorf("mismatch of board size and game cells: %w", domain.ErrBadInput)
	}

	var game generated.Game
	err = s.txRunner.WithTx(ctx, func(r repository.Repos) error {
		game, err = r.Games.Create(ctx, repository.CreateGameInput{
			PlayerID:  sessionUser.UserID,
			SessionID: sessionUser.SessionID,
			BoardID:   in.BoardID,
		})
		if err != nil {
			return fmt.Errorf("create game: %w", err)
		}

		var gameCells []repository.CreateGameCellItem
		gameCells, err = gameCellsToRepoItems(ctx, r.Cells, in.GameCells, board.ID)
		if err != nil {
			return fmt.Errorf("retrieve cells: %w", err)
		}

		_, err = r.GameCells.Create(ctx, repository.CreateGameCellsInput{
			GameID:    game.ID,
			GameCells: gameCells,
		})
		if err != nil {
			return fmt.Errorf("create game cells: %w", err)
		}

		return nil
	})
	if err != nil {
		return generated.Game{}, err
	}

	return game, nil
}

func (s *Games) UpdateStatus(ctx context.Context, in UpdateGameStatusInput) (generated.Game, error) {
	sessionUser, err := checkIfCallerOwnsGame(ctx, s.games, s.queries, in.GameID)
	if err != nil {
		return generated.Game{}, err
	}

	return s.games.UpdateStatus(ctx, repository.UpdateGameStatusInput{
		GameID:    in.GameID,
		Status:    in.Status,
		PlayerID:  sessionUser.UserID,
		SessionID: sessionUser.SessionID,
	})
}

func (s *Games) Delete(ctx context.Context, gameID pgtype.UUID) error {
	_, err := checkIfCallerOwnsGame(ctx, s.games, s.queries, gameID)
	if err != nil {
		return err
	}

	_, err = s.games.Delete(ctx, gameID)

	return err
}

func checkIfCallerOwnsGame(
	ctx context.Context,
	games *repository.Games,
	queries *generated.Queries,
	gameID pgtype.UUID,
) (generated.GetSessionUserByTokenRow, error) {
	sessionUser, err := middleware.RequireSession(ctx, queries)
	if err != nil {
		return generated.GetSessionUserByTokenRow{}, err
	}

	game, err := games.Get(ctx, gameID)
	if err != nil {
		return generated.GetSessionUserByTokenRow{}, err
	}

	// Authored games are owned by the player; anonymous games are owned by
	// the session that created them.
	if game.PlayerID.Valid {
		if sessionUser.UserID != game.PlayerID {
			return generated.GetSessionUserByTokenRow{}, domain.ErrForbidden
		}
	} else {
		if sessionUser.SessionID != game.SessionID {
			return generated.GetSessionUserByTokenRow{}, domain.ErrForbidden
		}
	}

	return sessionUser, nil
}

func gameCellsToRepoItems(ctx context.Context, cellRepo *repository.Cells, cells []CreateGameCellItem, boardID pgtype.UUID) ([]repository.CreateGameCellItem, error) {
	cellIDs := make([]pgtype.UUID, len(cells))
	for i, cell := range cells {
		cellIDs[i] = cell.CellID
	}

	boardCells, err := cellRepo.ListByIDs(ctx, cellIDs, boardID)
	if err != nil {
		return []repository.CreateGameCellItem{}, fmt.Errorf("get cells from board: %w", err)
	}

	cellMap := make(map[pgtype.UUID]*generated.Cell)
	for i := range boardCells {
		cellMap[boardCells[i].ID] = &boardCells[i]
	}

	result := make([]repository.CreateGameCellItem, len(cells))
	for i, cell := range cells {
		boardCell, exists := cellMap[cell.CellID]
		if !exists {
			return []repository.CreateGameCellItem{}, fmt.Errorf("cell not found: %w", domain.ErrBadInput)
		}

		result[i] = repository.CreateGameCellItem{
			CellID:   cell.CellID,
			Position: cell.Position,
			Content:  boardCell.Content,
		}
	}

	return result, nil
}
