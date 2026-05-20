package service

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/officeryoda/dozingo/internal/generated"
	"github.com/officeryoda/dozingo/internal/repository"
)

type Games interface {
	Get(ctx context.Context, id pgtype.UUID) (generated.Game, error)
	ListByPlayer(ctx context.Context, playerID pgtype.UUID) ([]generated.Game, error)
	ListByBoard(ctx context.Context, boardID pgtype.UUID) ([]generated.Game, error)
	Create(ctx context.Context, in repository.CreateGameInput) (generated.Game, error)
	UpdateStatus(ctx context.Context, in repository.UpdateGameStatusInput) (generated.Game, error)
	Delete(ctx context.Context, id pgtype.UUID) error
}

type games struct {
	repo *repository.Games
}

func NewGames(repo *repository.Games) Games {
	return &games{repo: repo}
}

func (s *games) Get(ctx context.Context, id pgtype.UUID) (generated.Game, error) {
	return s.repo.Get(ctx, id)
}

func (s *games) ListByPlayer(ctx context.Context, playerID pgtype.UUID) ([]generated.Game, error) {
	return s.repo.ListByPlayer(ctx, playerID)
}

func (s *games) ListByBoard(ctx context.Context, boardID pgtype.UUID) ([]generated.Game, error) {
	return s.repo.ListByBoard(ctx, boardID)
}

func (s *games) Create(ctx context.Context, in repository.CreateGameInput) (generated.Game, error) {
	// TODO(authz): once handlers pass the session, verify in.PlayerID matches
	// the authenticated user.
	return s.repo.Create(ctx, in)
}

func (s *games) UpdateStatus(ctx context.Context, in repository.UpdateGameStatusInput) (generated.Game, error) {
	// TODO(authz): verify the caller owns the game.
	return s.repo.UpdateStatus(ctx, in)
}

func (s *games) Delete(ctx context.Context, id pgtype.UUID) error {
	// TODO(authz): verify the caller owns the game.
	_, err := s.repo.Delete(ctx, id)
	return err
}
