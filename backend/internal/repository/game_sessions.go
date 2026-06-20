package repository

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/officeryoda/dozingo/internal/generated"
	"github.com/officeryoda/dozingo/internal/pgmap"
)

type GameSessions struct {
	queries *generated.Queries
}

func (r *GameSessions) Create(ctx context.Context, gameID pgtype.UUID) (generated.GameSession, error) {
	session, err := r.queries.CreateGameSession(ctx, gameID)
	if err != nil {
		return generated.GameSession{}, pgmap.TranslatePgErr(err)
	}

	return session, nil
}

func (r *GameSessions) UpdateHeartbeat(ctx context.Context, gameID pgtype.UUID) (session generated.GameSession, found bool, err error) {
	session, err = r.queries.UpdateHeartbeat(ctx, gameID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return generated.GameSession{}, false, nil
		}

		return generated.GameSession{}, false, pgmap.TranslatePgErr(err)
	}

	return session, true, nil
}

func (r *GameSessions) EndSessions(ctx context.Context, gameID pgtype.UUID) (int64, error) {
	n, err := r.queries.EndGameSessions(ctx, gameID)
	if err != nil {
		return 0, pgmap.TranslatePgErr(err)
	}

	return n, nil
}

func (r *GameSessions) CloseStaleSessions(ctx context.Context, timeout time.Duration) (int64, error) {
	n, err := r.queries.CloseStaleSessions(ctx, pgmap.PgIntervalFromDuration(&timeout))
	if err != nil {
		return 0, pgmap.TranslatePgErr(err)
	}

	return n, nil
}

func (r *GameSessions) HasOpenSession(ctx context.Context, gameID pgtype.UUID) (bool, error) {
	hasOpen, err := r.queries.HasOpenGameSession(ctx, gameID)
	if err != nil {
		return false, pgmap.TranslatePgErr(err)
	}

	return hasOpen, nil
}

func (r *GameSessions) GetPlaytimeByGame(ctx context.Context, gameID pgtype.UUID) (int64, error) {
	seconds, err := r.queries.GetPlaytimeByGame(ctx, gameID)
	if err != nil {
		return 0, pgmap.TranslatePgErr(err)
	}

	return seconds, nil
}

func (r *GameSessions) GetPlaytimeByBoard(ctx context.Context, boardID pgtype.UUID) (int64, error) {
	seconds, err := r.queries.GetPlaytimeByBoard(ctx, boardID)
	if err != nil {
		return 0, pgmap.TranslatePgErr(err)
	}

	return seconds, nil
}

func (r *GameSessions) GetPlaytimeByPlayer(ctx context.Context, playerID pgtype.UUID) (int64, error) {
	seconds, err := r.queries.GetPlaytimeByPlayer(ctx, playerID)
	if err != nil {
		return 0, pgmap.TranslatePgErr(err)
	}

	return seconds, nil
}

func (r *GameSessions) GetTotalPlaytime(ctx context.Context, period time.Duration) (int64, error) {
	seconds, err := r.queries.GetTotalPlaytime(ctx, pgmap.PgIntervalFromDuration(&period))
	if err != nil {
		return 0, pgmap.TranslatePgErr(err)
	}

	return seconds, nil
}
