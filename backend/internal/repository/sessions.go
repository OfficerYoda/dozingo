package repository

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/officeryoda/dozingo/internal/generated"
	"github.com/officeryoda/dozingo/internal/pgmap"
)

// All token-typed parameters and fields in this package are SHA-256 hex
// digests, never plaintext. Callers (middleware, service layer) are
// responsible for hashing the cookie value via auth.HashToken before
// invoking any of these methods. Token-typed fields on returned rows
// likewise hold the hash and can be passed back to other methods as-is.

type Sessions struct {
	queries *generated.Queries
}

type CreateSessionInput struct {
	UserID    pgtype.UUID
	TokenHash string
	ExpiresAt time.Time
}

func (r *Sessions) Create(ctx context.Context, in CreateSessionInput) (generated.Session, error) {
	session, err := r.queries.CreateSession(ctx, generated.CreateSessionParams{
		UserID:    in.UserID,
		Token:     in.TokenHash,
		ExpiresAt: pgmap.PgTimestamptzFromTime(&in.ExpiresAt),
	})
	if err != nil {
		return generated.Session{}, pgmap.TranslatePgErr(err)
	}

	return session, nil
}

func (r *Sessions) Extend(ctx context.Context, tokenHash string, expiresAt time.Time) (generated.Session, error) {
	session, err := r.queries.ExtendSessionByToken(ctx, generated.ExtendSessionByTokenParams{
		Token:     tokenHash,
		ExpiresAt: pgmap.PgTimestamptzFromTime(&expiresAt),
	})
	if err != nil {
		return generated.Session{}, pgmap.TranslatePgErr(err)
	}

	return session, nil
}

func (r *Sessions) AttachUser(ctx context.Context, tokenHash string, userID pgtype.UUID) (generated.Session, error) {
	session, err := r.queries.AttachUserToSession(ctx, generated.AttachUserToSessionParams{
		Token:  tokenHash,
		UserID: userID,
	})
	if err != nil {
		return generated.Session{}, pgmap.TranslatePgErr(err)
	}

	return session, nil
}

func (r *Sessions) SetTwoFAPending(ctx context.Context, tokenHash string, status bool) (generated.Session, error) {
	session, err := r.queries.SetTwoFAPending(ctx, generated.SetTwoFAPendingParams{
		Token:        tokenHash,
		TwoFaPending: status,
	})
	if err != nil {
		return generated.Session{}, pgmap.TranslatePgErr(err)
	}

	return session, nil
}

func (r *Sessions) Delete(ctx context.Context, tokenHash string) error {
	err := r.queries.DeleteSessionByToken(ctx, tokenHash)
	if err != nil {
		return pgmap.TranslatePgErr(err)
	}

	return nil
}

func (r *Sessions) DeleteByUserID(ctx context.Context, userID pgtype.UUID) error {
	err := r.queries.DeleteSessionByUserID(ctx, userID)
	if err != nil {
		return pgmap.TranslatePgErr(err)
	}

	return nil
}

func (r *Sessions) DeleteOtherSessionsFromUser(ctx context.Context, userID, sessionID pgtype.UUID) error {
	err := r.queries.DeleteOtherSessionsFromUser(ctx, generated.DeleteOtherSessionsFromUserParams{
		UserID:    userID,
		SessionID: sessionID,
	})
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

func (r *Sessions) GetUserByToken(ctx context.Context, tokenHash string) (generated.GetSessionUserByTokenRow, error) {
	user, err := r.queries.GetSessionUserByToken(ctx, tokenHash)
	if err != nil {
		return generated.GetSessionUserByTokenRow{}, pgmap.TranslatePgErr(err)
	}

	return user, nil
}
