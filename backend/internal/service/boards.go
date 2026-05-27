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
}

type CreateBoardInput struct {
	Title       string
	Description *string
	Size        int32
}

func (s *Boards) List(ctx context.Context, filter BoardListFilter) ([]generated.Board, error) {
	return s.boards.List(ctx, repository.BoardListFilter{
		AuthorID: filter.AuthorID,
		Size:     filter.Size,
		Sort:     filter.Sort,
		Limit:    filter.Limit,
	})
}

func (s *Boards) Get(ctx context.Context, boardID pgtype.UUID) (generated.Board, error) {
	return s.boards.Get(ctx, boardID)
}

func (s *Boards) Create(ctx context.Context, in CreateBoardInput) (generated.Board, error) {
	sessionUser, err := middleware.RequireSession(ctx, s.queries)
	if err != nil {
		return generated.Board{}, fmt.Errorf("session required: %w", err)
	}
	if !sessionUser.UserID.Valid {
		return generated.Board{}, fmt.Errorf("authenticated user required: %w", domain.ErrUnauthorized)
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
