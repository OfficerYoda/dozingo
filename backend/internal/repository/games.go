package repository

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/officeryoda/dozingo/internal/generated"
)

type Games struct {
	queries *generated.Queries
}

type CreateGameInput struct {
	PlayerID pgtype.UUID
	BoardID  pgtype.UUID
}

type UpdateGameStatusInput struct {
	GameID   pgtype.UUID
	PlayerID pgtype.UUID
	Status   string
}

func (r *Games) Get(ctx context.Context, id pgtype.UUID) (generated.Game, error) {
	game, err := r.queries.GetGameByID(ctx, id)
	if err != nil {
		return generated.Game{}, translatePgErr(err)
	}
	return game, nil
}

func (r *Games) ListByPlayer(ctx context.Context, playerID pgtype.UUID) ([]generated.Game, error) {
	games, err := r.queries.ListGamesByPlayer(ctx, playerID)
	if err != nil {
		return []generated.Game{}, translatePgErr(err)
	}
	return games, nil
}

func (r *Games) ListByBoard(ctx context.Context, boardID pgtype.UUID) ([]generated.Game, error) {
	games, err := r.queries.ListGamesByBoard(ctx, boardID)
	if err != nil {
		return []generated.Game{}, translatePgErr(err)
	}
	return games, nil
}

func (r *Games) Create(ctx context.Context, in CreateGameInput) (generated.Game, error) {
	game, err := r.queries.CreateGame(ctx, generated.CreateGameParams{
		PlayerID: in.PlayerID,
		BoardID:  in.BoardID,
	})
	if err != nil {
		return generated.Game{}, translatePgErr(err)
	}
	return game, nil
}

func (r *Games) UpdateStatus(ctx context.Context, in UpdateGameStatusInput) (generated.Game, error) {
	game, err := r.queries.UpdateGameStatus(ctx, generated.UpdateGameStatusParams{
		ID:       in.GameID,
		PlayerID: in.PlayerID,
		Status:   in.Status,
	})
	if err != nil {
		return generated.Game{}, translatePgErr(err)
	}
	return game, nil
}

func (r *Games) Delete(ctx context.Context, id pgtype.UUID) (generated.Game, error) {
	game, err := r.queries.DeleteGame(ctx, id)
	if err != nil {
		return generated.Game{}, translatePgErr(err)
	}
	return game, nil
}
