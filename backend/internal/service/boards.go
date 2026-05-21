package service

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/officeryoda/dozingo/internal/generated"
	"github.com/officeryoda/dozingo/internal/repository"
)

type Boards struct {
	repo *repository.Boards
}

func NewBoards(repo *repository.Boards) *Boards {
	return &Boards{repo: repo}
}

func (s *Boards) List(ctx context.Context, filter repository.BoardListFilter) ([]generated.Board, error) {
	return s.repo.List(ctx, filter)
}

func (s *Boards) Get(ctx context.Context, id pgtype.UUID) (generated.Board, error) {
	return s.repo.Get(ctx, id)
}

func (s *Boards) Create(ctx context.Context, in repository.CreateBoardInput) (generated.Board, error) {
	// TODO(authz): once handlers pass the session, verify in.AuthorID matches
	// the authenticated user (or admin).
	return s.repo.Create(ctx, in)
}

func (s *Boards) Delete(ctx context.Context, id pgtype.UUID) error {
	// TODO(authz): once handlers pass the session, verify the caller owns the
	// board before deleting.
	_, err := s.repo.Delete(ctx, id)
	return err
}
