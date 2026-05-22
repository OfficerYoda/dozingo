package service

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/officeryoda/dozingo/internal/generated"
	"github.com/officeryoda/dozingo/internal/repository"
)

type Games struct {
	games *repository.Games
}

func NewGames(games *repository.Games) *Games {
	return &Games{games: games}
}

func (s *Games) Get(ctx context.Context, id pgtype.UUID) (generated.Game, error) {
	return s.games.Get(ctx, id)
}

func (s *Games) ListByPlayer(ctx context.Context, playerID pgtype.UUID) ([]generated.Game, error) {
	return s.games.ListByPlayer(ctx, playerID)
}

func (s *Games) ListByBoard(ctx context.Context, boardID pgtype.UUID) ([]generated.Game, error) {
	return s.games.ListByBoard(ctx, boardID)
}

func (s *Games) Create(ctx context.Context, in repository.CreateGameInput) (generated.Game, error) {
	// TODO(authz): once handlers pass the session, verify in.PlayerID matches
	// the authenticated user.
	return s.games.Create(ctx, in)
}

func (s *Games) UpdateStatus(ctx context.Context, in repository.UpdateGameStatusInput) (generated.Game, error) {
	// TODO(authz): verify the caller owns the game.
	return s.games.UpdateStatus(ctx, in)
}

func (s *Games) Delete(ctx context.Context, id pgtype.UUID) error {
	// TODO(authz): verify the caller owns the game.
	_, err := s.games.Delete(ctx, id)
	return err
}
