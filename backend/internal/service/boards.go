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

type Boards struct {
	boards  *repository.Boards
	queries *generated.Queries
}

func NewBoards(boards *repository.Boards, queries *generated.Queries) *Boards {
	return &Boards{boards: boards, queries: queries}
}

type BoardListFilter struct {
	AuthorID string
	Size     int32
	Sort     string
	Limit    int32
	Search   string
}

type CreateBoardInput struct {
	Title       string
	Description *string
	Size        int32
}

func (s *Boards) ListBySession(ctx context.Context, filter BoardListFilter) ([]repository.BoardWithStats, error) {
	sessionUser, err := requiresSessionUser(ctx, s.queries)
	if err != nil {
		return []repository.BoardWithStats{}, err
	}

	return s.List(ctx, BoardListFilter{
		AuthorID: sessionUser.UserID.String(),
		Size:     filter.Size,
		Sort:     filter.Sort,
		Limit:    filter.Limit,
	})
}

func (s *Boards) List(ctx context.Context, filter BoardListFilter) ([]repository.BoardWithStats, error) {
	return s.boards.List(ctx, repository.BoardListFilter{
		AuthorID: filter.AuthorID,
		Size:     filter.Size,
		Sort:     filter.Sort,
		Limit:    filter.Limit,
		Search:   filter.Search,
	})
}

func (s *Boards) Get(ctx context.Context, boardID pgtype.UUID) (repository.BoardWithStats, error) {
	return s.boards.Get(ctx, boardID)
}

func (s *Boards) Create(ctx context.Context, in CreateBoardInput) (repository.BoardWithStats, error) {
	sessionUser, err := middleware.RequireSession(ctx, s.queries)
	if err != nil {
		return repository.BoardWithStats{}, fmt.Errorf("session required: %w", err)
	}
	if !sessionUser.UserID.Valid {
		return repository.BoardWithStats{}, fmt.Errorf("authenticated user required: %w", domain.ErrUnauthorized)
	}

	return s.boards.Create(ctx, repository.CreateBoardInput{
		Title:       in.Title,
		Description: in.Description,
		Size:        in.Size,
		AuthorID:    sessionUser.UserID,
	})
}

func (s *Boards) Delete(ctx context.Context, boardID pgtype.UUID) error {
	sessionUser, err := requiresSessionUser(ctx, s.queries)
	if err != nil {
		return err
	}

	board, err := s.boards.Get(ctx, boardID)
	if err != nil {
		return err
	}
	if board.AuthorID != sessionUser.UserID {
		return domain.ErrForbidden
	}

	_, err = s.boards.Delete(ctx, boardID)
	return err
}

func (s *Boards) TotalGamesPlayed(ctx context.Context, boardID pgtype.UUID) (generated.GetTotalGamesPlayedForBoardRow, error) {
	playedGames, err := s.boards.TotalGamesPlayed(ctx, boardID)
	return playedGames, err
}
