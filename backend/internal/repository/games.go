package repository

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/officeryoda/dozingo/internal/generated"
	"github.com/officeryoda/dozingo/internal/pgmap"
)

type Games struct {
	queries *generated.Queries
}

type CreateGameInput struct {
	PlayerID  pgtype.UUID
	SessionID pgtype.UUID
	BoardID   pgtype.UUID
}

type UpdateGameStatusInput struct {
	GameID    pgtype.UUID
	PlayerID  pgtype.UUID
	SessionID pgtype.UUID
	Status    string
}

func (r *Games) Get(ctx context.Context, gameID pgtype.UUID) (generated.Game, error) {
	game, err := r.queries.GetGameByID(ctx, gameID)
	if err != nil {
		return generated.Game{}, pgmap.TranslatePgErr(err)
	}

	return game, nil
}

func (r *Games) ListByPlayer(ctx context.Context, playerID pgtype.UUID) ([]generated.Game, error) {
	games, err := r.queries.ListGamesByPlayer(ctx, playerID)
	if err != nil {
		return []generated.Game{}, pgmap.TranslatePgErr(err)
	}

	return games, nil
}

func (r *Games) ListBySession(ctx context.Context, sessionID pgtype.UUID) ([]generated.Game, error) {
	games, err := r.queries.ListGamesBySession(ctx, sessionID)
	if err != nil {
		return []generated.Game{}, pgmap.TranslatePgErr(err)
	}

	return games, nil
}

func (r *Games) ListByBoard(ctx context.Context, boardID pgtype.UUID) ([]generated.Game, error) {
	games, err := r.queries.ListGamesByBoard(ctx, boardID)
	if err != nil {
		return []generated.Game{}, pgmap.TranslatePgErr(err)
	}

	return games, nil
}

func (r *Games) Create(ctx context.Context, in CreateGameInput) (generated.Game, error) {
	game, err := r.queries.CreateGame(ctx, generated.CreateGameParams{
		PlayerID:  in.PlayerID,
		SessionID: in.SessionID,
		BoardID:   in.BoardID,
	})
	if err != nil {
		return generated.Game{}, pgmap.TranslatePgErr(err)
	}

	return game, nil
}

func (r *Games) UpdateStatus(ctx context.Context, in UpdateGameStatusInput) (generated.Game, error) {
	game, err := r.queries.UpdateGameStatus(ctx, generated.UpdateGameStatusParams{
		GameID:    in.GameID,
		PlayerID:  in.PlayerID,
		SessionID: in.SessionID,
		Status:    generated.GameStatus(in.Status),
	})
	if err != nil {
		return generated.Game{}, pgmap.TranslatePgErr(err)
	}

	return game, nil
}

func (r *Games) Delete(ctx context.Context, gameID pgtype.UUID) (generated.Game, error) {
	game, err := r.queries.DeleteGame(ctx, gameID)
	if err != nil {
		return generated.Game{}, pgmap.TranslatePgErr(err)
	}

	return game, nil
}

func (r *Games) SetBingoCount(ctx context.Context, gameID pgtype.UUID, count int32) (generated.Game, error) {
	game, err := r.queries.SetBingoCount(ctx, generated.SetBingoCountParams{
		GameID:     gameID,
		BingoCount: count,
	})
	if err != nil {
		return generated.Game{}, pgmap.TranslatePgErr(err)
	}

	return game, nil
}

// AbandonInactive returns the number of games affected.
func (r *Games) AbandonInactive(ctx context.Context, timeout time.Duration) (int64, error) {
	n, err := r.queries.AbandonInactiveGames(ctx, pgmap.PgIntervalFromDuration(&timeout))
	if err != nil {
		return 0, pgmap.TranslatePgErr(err)
	}

	return n, nil
}
