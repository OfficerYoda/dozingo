package repository

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/officeryoda/dozingo/internal/generated"
	"github.com/officeryoda/dozingo/internal/pgmap"
)

type Sessions struct {
	queries *generated.Queries
}

type CreateSessionInput struct {
	UserID    pgtype.UUID
	Token     string
	ExpiresAt time.Time
}

func (r *Sessions) Create(ctx context.Context, in CreateSessionInput) (generated.Session, error) {
	session, err := r.queries.CreateSession(ctx, generated.CreateSessionParams{
		UserID:    in.UserID,
		Token:     in.Token,
		ExpiresAt: pgmap.PgTimestamptzFromTime(in.ExpiresAt),
	})
	if err != nil {
		return generated.Session{}, pgmap.TranslatePgErr(err)
	}
	return session, nil
}

func (r *Sessions) Extend(ctx context.Context, token string, expiresAt time.Time) (generated.Session, error) {
	session, err := r.queries.ExtendSessionByToken(ctx, generated.ExtendSessionByTokenParams{
		Token:     token,
		ExpiresAt: pgmap.PgTimestamptzFromTime(expiresAt),
	})
	if err != nil {
		return generated.Session{}, pgmap.TranslatePgErr(err)
	}
	return session, nil
}

func (r *Sessions) AttachUser(ctx context.Context, token string, userID pgtype.UUID) (generated.Session, error) {
	session, err := r.queries.AttachUserToSession(ctx, generated.AttachUserToSessionParams{
		Token:  token,
		UserID: userID,
	})
	if err != nil {
		return generated.Session{}, pgmap.TranslatePgErr(err)
	}
	return session, nil
}

func (r *Sessions) Delete(ctx context.Context, token string) error {
	err := r.queries.DeleteSessionByToken(ctx, token)
	if err != nil {
		return pgmap.TranslatePgErr(err)
	}
	return nil
}

func (r *Sessions) DeleteExpiredSessions(ctx context.Context) error {
	err := r.queries.DeleteExpiredSessions(ctx)
	if err != nil {
		return pgmap.TranslatePgErr(err)
	}
	return nil
}

func (r *Sessions) GetUserByToken(ctx context.Context, token string) (generated.GetSessionUserByTokenRow, error) {
	user, err := r.queries.GetSessionUserByToken(ctx, token)
	if err != nil {
		return generated.GetSessionUserByTokenRow{}, pgmap.TranslatePgErr(err)
	}
	return user, nil
}
