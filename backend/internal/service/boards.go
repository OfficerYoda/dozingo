package service

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/officeryoda/dozingo/internal/generated"
	"github.com/officeryoda/dozingo/internal/repository"
)

type Boards struct {
	boards *repository.Boards
}

func NewBoards(boards *repository.Boards) *Boards {
	return &Boards{boards: boards}
}

func (s *Boards) List(ctx context.Context, filter repository.BoardListFilter) ([]generated.Board, error) {
	return s.boards.List(ctx, filter)
}

func (s *Boards) Get(ctx context.Context, id pgtype.UUID) (generated.Board, error) {
	return s.boards.Get(ctx, id)
}

func (s *Boards) Create(ctx context.Context, in repository.CreateBoardInput) (generated.Board, error) {
	// TODO(authz): once handlers pass the session, verify in.AuthorID matches
	// the authenticated user (or admin).
	return s.boards.Create(ctx, in)
}

func (s *Boards) Delete(ctx context.Context, id pgtype.UUID) error {
	// TODO(authz): once handlers pass the session, verify the caller owns the
	// board before deleting.
	_, err := s.boards.Delete(ctx, id)
	return err
}
