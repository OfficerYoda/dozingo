package repository

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/officeryoda/dozingo/internal/generated"
	"github.com/officeryoda/dozingo/internal/pgmap"
)

type VerificationTokens struct {
	queries *generated.Queries
}

type CreateVerificationTokenInput struct {
	UserID    pgtype.UUID
	Token     string
	TokenType string
	ExpiresAt time.Time
}

type GetByTokenForUserInput struct {
	UserID    pgtype.UUID
	TokenType string
}

func (r *VerificationTokens) Create(ctx context.Context, in CreateVerificationTokenInput) (generated.VerificationToken, error) {
	tokenType := generated.TokenType(in.TokenType)
	token, err := r.queries.CreateVerificationToken(ctx, generated.CreateVerificationTokenParams{
		UserID:    in.UserID,
		Token:     in.Token,
		Type:      tokenType,
		ExpiresAt: pgmap.PgTimestamptzFromTime(&in.ExpiresAt),
	})
	if err != nil {
		return generated.VerificationToken{}, pgmap.TranslatePgErr(err)
	}

	return token, nil
}

func (r *VerificationTokens) GetByToken(ctx context.Context, token string) (generated.VerificationToken, error) {
	verificationToken, err := r.queries.GetVerificationTokenByToken(ctx, token)
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

func (r *VerificationTokens) Delete(ctx context.Context, token string) error {
	err := r.queries.DeleteVerificationToken(ctx, token)
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
