package repository

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/officeryoda/dozingo/internal/generated"
	"github.com/officeryoda/dozingo/internal/pgmap"
)

// All token-typed parameters and fields in this package are SHA-256 hex
// digests, never plaintext. Callers (the service layer) are responsible
// for hashing the user-supplied token via auth.HashToken before invoking
// any of these methods. Token-typed fields on returned rows likewise
// hold the hash and can be passed back to other methods as-is.

type VerificationTokens struct {
	queries *generated.Queries
}

type CreateVerificationTokenInput struct {
	UserID    pgtype.UUID
	TokenHash string
	TokenType generated.TokenType
	ExpiresAt time.Time
}

type GetByTokenForUserInput struct {
	UserID    pgtype.UUID
	TokenType generated.TokenType
}

func (r *VerificationTokens) Create(ctx context.Context, in CreateVerificationTokenInput) (generated.VerificationToken, error) {
	token, err := r.queries.CreateVerificationToken(ctx, generated.CreateVerificationTokenParams{
		UserID:    in.UserID,
		Token:     in.TokenHash,
		Type:      in.TokenType,
		ExpiresAt: pgmap.PgTimestamptzFromTime(&in.ExpiresAt),
	})
	if err != nil {
		return generated.VerificationToken{}, pgmap.TranslatePgErr(err)
	}

	return token, nil
}

func (r *VerificationTokens) GetByToken(ctx context.Context, tokenHash string) (generated.VerificationToken, error) {
	verificationToken, err := r.queries.GetVerificationTokenByToken(ctx, tokenHash)
	if err != nil {
		return generated.VerificationToken{}, pgmap.TranslatePgErr(err)
	}

	return verificationToken, nil
}

func (r *VerificationTokens) GetValidTokenForUser(ctx context.Context, in GetByTokenForUserInput) (generated.VerificationToken, error) {
	tokenType := generated.TokenType(in.TokenType)
	token, err := r.queries.GetValidTokenForUser(ctx, generated.GetValidTokenForUserParams{
		UserID: in.UserID,
		Type:   tokenType,
	})
	if err != nil {
		return generated.VerificationToken{}, pgmap.TranslatePgErr(err)
	}

	return token, nil
}

func (r *VerificationTokens) Delete(ctx context.Context, tokenHash string) error {
	err := r.queries.DeleteVerificationToken(ctx, tokenHash)
	if err != nil {
		return pgmap.TranslatePgErr(err)
	}

	return nil
}

func (r *VerificationTokens) DeleteExpired(ctx context.Context) error {
	err := r.queries.DeleteExpiredTokens(ctx)
	if err != nil {
		return pgmap.TranslatePgErr(err)
	}

	return nil
}
