package repository

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/officeryoda/dozingo/internal/generated"
)

type UserPasswords struct {
	queries *generated.Queries
}

func (r *UserPasswords) GetHashForUserID(ctx context.Context, userID pgtype.UUID) (string, error) {
	hash, err := r.queries.GetPasswordHashByUserID(ctx, userID)
	if err != nil {
		return "", translatePgErr(err)
	}
	return hash, nil
}

func (r *UserPasswords) Upsert(ctx context.Context, userID pgtype.UUID, passwordHash string) (generated.UserPassword, error) {
	userPassword, err := r.queries.UpsertUserPassword(ctx, generated.UpsertUserPasswordParams{
		UserID:       userID,
		PasswordHash: passwordHash,
	})
	if err != nil {
		return generated.UserPassword{}, translatePgErr(err)
	}
	return userPassword, nil
}
