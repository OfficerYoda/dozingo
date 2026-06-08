package repository

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/officeryoda/dozingo/internal/generated"
	"github.com/officeryoda/dozingo/internal/pgmap"
)

type UserPasswords struct {
	queries *generated.Queries
}

func (r *UserPasswords) GetHashForUserID(ctx context.Context, userID pgtype.UUID) (string, error) {
	hash, err := r.queries.GetPasswordHashByUserID(ctx, userID)
	if err != nil {
		return "", pgmap.TranslatePgErr(err)
	}

	return hash, nil
}

func (r *UserPasswords) Upsert(ctx context.Context, userID pgtype.UUID, passwordHash string) (generated.UserPassword, error) {
	userPassword, err := r.queries.UpsertUserPassword(ctx, generated.UpsertUserPasswordParams{
		UserID:       userID,
		PasswordHash: passwordHash,
	})
	if err != nil {
		return generated.UserPassword{}, pgmap.TranslatePgErr(err)
	}

	return userPassword, nil
}
