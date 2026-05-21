package service

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/officeryoda/dozingo/internal/generated"
	"github.com/officeryoda/dozingo/internal/repository"
)

type Cells struct {
	repo *repository.Cells
}

func NewCells(repo *repository.Cells) *Cells {
	return &Cells{repo: repo}
}

func (s *Cells) ListByBoardID(ctx context.Context, boardID pgtype.UUID) ([]generated.Cell, error) {
	return s.repo.ListByBoardID(ctx, boardID)
}

func (s *Cells) Create(ctx context.Context, in repository.CreateCellInput) (generated.Cell, error) {
	// TODO(authz): once handlers pass the session, verify the caller owns
	// the board before creating cells on it.
	if in.Value == nil {
		def := int32(1)
		in.Value = &def
	}
	return s.repo.Create(ctx, in)
}

func (s *Cells) Update(ctx context.Context, in repository.UpdateCellInput) (generated.Cell, error) {
	// TODO(authz): verify the caller owns the board.
	return s.repo.Update(ctx, in)
}

func (s *Cells) Delete(ctx context.Context, cellID, boardID pgtype.UUID) error {
	// TODO(authz): verify the caller owns the board.
	_, err := s.repo.Delete(ctx, cellID, boardID)
	return err
}
