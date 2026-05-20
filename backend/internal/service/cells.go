package service

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/officeryoda/dozingo/internal/generated"
	"github.com/officeryoda/dozingo/internal/repository"
)

type Cells interface {
	ListByBoardID(ctx context.Context, boardID pgtype.UUID) ([]generated.Cell, error)
	Create(ctx context.Context, in repository.CreateCellInput) (generated.Cell, error)
	Update(ctx context.Context, in repository.UpdateCellInput) (generated.Cell, error)
	Delete(ctx context.Context, cellID, boardID pgtype.UUID) error
}

type cells struct {
	repo *repository.Cells
}

func NewCells(repo *repository.Cells) Cells {
	return &cells{repo: repo}
}

func (s *cells) ListByBoardID(ctx context.Context, boardID pgtype.UUID) ([]generated.Cell, error) {
	return s.repo.ListByBoardID(ctx, boardID)
}

func (s *cells) Create(ctx context.Context, in repository.CreateCellInput) (generated.Cell, error) {
	// TODO(authz): once handlers pass the session, verify the caller owns
	// the board before creating cells on it.
	if in.Value == nil {
		def := int32(1)
		in.Value = &def
	}
	return s.repo.Create(ctx, in)
}

func (s *cells) Update(ctx context.Context, in repository.UpdateCellInput) (generated.Cell, error) {
	// TODO(authz): verify the caller owns the board.
	return s.repo.Update(ctx, in)
}

func (s *cells) Delete(ctx context.Context, cellID, boardID pgtype.UUID) error {
	// TODO(authz): verify the caller owns the board.
	_, err := s.repo.Delete(ctx, cellID, boardID)
	return err
}
