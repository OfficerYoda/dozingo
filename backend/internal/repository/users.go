package repository

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/officeryoda/dozingo/internal/generated"
)

type Users struct {
	queries *generated.Queries
}

func (r *Users) GetForPasswordLogin(ctx context.Context, username string) (generated.GetUserForPasswordLoginRow, error) {
	user, err := r.queries.GetUserForPasswordLogin(ctx, username)
	if err != nil {
		return generated.GetUserForPasswordLoginRow{}, translatePgErr(err)
	}
	return user, nil
}

func (r *Users) GetByID(ctx context.Context, userID pgtype.UUID) (generated.User, error) {
	user, err := r.queries.GetUserByID(ctx, userID)
	if err != nil {
		return generated.User{}, translatePgErr(err)
	}
	return user, nil
}

func (r *Users) GetByUsername(ctx context.Context, username string) (generated.User, error) {
	user, err := r.queries.GetUserByUsername(ctx, username)
	if err != nil {
		return generated.User{}, translatePgErr(err)
	}
	return user, nil
}

func (r *Users) Create(ctx context.Context, username string, email *string) (generated.User, error) {
	user, err := r.queries.CreateUser(ctx, generated.CreateUserParams{
		Username: username,
		Email:    pgTextFromString(email),
	})
	if err != nil {
		return generated.User{}, translatePgErr(err)
	}
	return user, nil
}

func (r *Users) Delete(ctx context.Context, userID pgtype.UUID) (generated.User, error) {
	user, err := r.queries.DeleteUser(ctx, userID)
	if err != nil {
		return generated.User{}, translatePgErr(err)
	}
	return user, nil
}
