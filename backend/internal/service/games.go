package service

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/officeryoda/dozingo/internal/domain"
	"github.com/officeryoda/dozingo/internal/generated"
	"github.com/officeryoda/dozingo/internal/repository"
)

type Games struct {
	games   *repository.Games
	queries *generated.Queries
}

func NewGames(games *repository.Games, queries *generated.Queries) *Games {
	return &Games{games: games, queries: queries}
}

type UpdateGameStatusInput struct {
	GameID pgtype.UUID
	Status string
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

func (s *Games) Create(ctx context.Context, boardID pgtype.UUID) (generated.Game, error) {
	sessionUser, err := requiresSessionUser(ctx, s.queries)
	if err != nil {
		return generated.Game{}, nil
	}

	return s.games.Create(ctx, repository.CreateGameInput{
		BoardID:  boardID,
		PlayerID: sessionUser.UserID,
	})
}

func (s *Games) UpdateStatus(ctx context.Context, in UpdateGameStatusInput) (generated.Game, error) {
	sessionUser, err := checkIfCallerOwnsGame(ctx, s, in.GameID)
	if err != nil {
		return generated.Game{}, err
	}

	return s.games.UpdateStatus(ctx, repository.UpdateGameStatusInput{
		GameID:   in.GameID,
		Status:   in.Status,
		PlayerID: sessionUser.UserID,
	})
}

func (s *Games) Delete(ctx context.Context, gameID pgtype.UUID) error {
	_, err := checkIfCallerOwnsGame(ctx, s, gameID)
	if err != nil {
		return err
	}

	_, err = s.games.Delete(ctx, gameID)
	return err
}

func checkIfCallerOwnsGame(ctx context.Context, s *Games, gameID pgtype.UUID) (generated.GetSessionUserByTokenRow, error) {
	sessionUser, err := requiresSessionUser(ctx, s.queries)
	if err != nil {
		return generated.GetSessionUserByTokenRow{}, nil
	}

	game, err := s.games.Get(ctx, gameID)
	if err != nil {
		return generated.GetSessionUserByTokenRow{}, err
	}

	if sessionUser.UserID != game.PlayerID {
		return generated.GetSessionUserByTokenRow{}, domain.ErrUnauthorized
	}
	return sessionUser, nil
}
