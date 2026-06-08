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

type Cells struct {
	cells   *repository.Cells
	boards  *repository.Boards
	queries *generated.Queries
}

func NewCells(cells *repository.Cells, boards *repository.Boards, queries *generated.Queries) *Cells {
	return &Cells{
		cells:   cells,
		boards:  boards,
		queries: queries,
	}
}

type CreateCellInput struct {
	BoardID pgtype.UUID
	Content string
	Value   *int32
}

type UpdateCellInput struct {
	CellID  pgtype.UUID
	BoardID pgtype.UUID
	Content *string
	Value   *int32
}

type DeleteCellInput struct {
	CellID  pgtype.UUID
	BoardID pgtype.UUID
}

func (s *Cells) ListByBoardID(ctx context.Context, boardID pgtype.UUID) ([]generated.Cell, error) {
	return s.cells.ListByBoardID(ctx, boardID)
}

func (s *Cells) Create(ctx context.Context, in CreateCellInput) (generated.Cell, error) {
	err := checkIfCallerOwnsBoard(ctx, s, in.BoardID)
	if err != nil {
		return generated.Cell{}, err
	}

	cellValue := in.Value
	if cellValue == nil {
		def := int32(1)
		cellValue = &def
	}

	return s.cells.Create(ctx, repository.CreateCellInput{
		BoardID: in.BoardID,
		Content: in.Content,
		Value:   *cellValue,
	})
}

func (s *Cells) Update(ctx context.Context, in UpdateCellInput) (generated.Cell, error) {
	err := checkIfCallerOwnsBoard(ctx, s, in.BoardID)
	if err != nil {
		return generated.Cell{}, err
	}

	if err := s.assertCellOnBoard(ctx, in.CellID, in.BoardID); err != nil {
		return generated.Cell{}, err
	}

	return s.cells.Update(ctx, repository.UpdateCellInput(in))
}

func (s *Cells) Delete(ctx context.Context, in DeleteCellInput) error {
	err := checkIfCallerOwnsBoard(ctx, s, in.BoardID)
	if err != nil {
		return err
	}

	if err := s.assertCellOnBoard(ctx, in.CellID, in.BoardID); err != nil {
		return err
	}

	_, err = s.cells.Delete(ctx, repository.DeleteCellInput(in))
	return err
}

// assertCellOnBoard fetches the cell by its id and verifies it really
// belongs to the board the caller named in the URL. The repository's
// UpdateCell/DeleteCell SQL already scopes to (cell_id, board_id) in
// the WHERE clause and would reject a cross-board id with no-rows. This
// extra check is defense-in-depth: it surfaces the error as an explicit
// "cell not found on this board" before the mutation runs and protects
// against a future SQL regression that drops the board_id scoping.
func (s *Cells) assertCellOnBoard(ctx context.Context, cellID, boardID pgtype.UUID) error {
	cell, err := s.cells.GetByID(ctx, cellID)
	if err != nil {
		return err
	}
	if cell.BoardID != boardID {
		return fmt.Errorf("cell does not belong to board: %w", domain.ErrNotFound)
	}
	return nil
}

func checkIfCallerOwnsBoard(ctx context.Context, s *Cells, boardID pgtype.UUID) error {
	sessionUser, err := middleware.RequireSession(ctx, s.queries)
	if err != nil {
		return fmt.Errorf("session required: %w", err)
	}

	board, err := s.boards.Get(ctx, boardID)
	if err != nil {
		return err
	}
	if board.AuthorID != sessionUser.UserID {
		return fmt.Errorf("caller doesn't own board: %w", domain.ErrForbidden)
	}

	return nil
}
