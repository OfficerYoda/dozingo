package repository

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/officeryoda/dozingo/internal/generated"
	"github.com/officeryoda/dozingo/internal/pgmap"
)

type TwoFactor struct {
	queries *generated.Queries
}

func (r *TwoFactor) GetByUserID(ctx context.Context, userID pgtype.UUID) (generated.UserTwoFactor, error) {
	twoFactor, err := r.queries.GetTwoFactorByUserID(ctx, userID)
	if err != nil {
		return generated.UserTwoFactor{}, pgmap.TranslatePgErr(err)
	}

	return twoFactor, nil
}

func (r *TwoFactor) Create(ctx context.Context, userID pgtype.UUID, encryptedSecret string) (generated.UserTwoFactor, error) {
	twoFactor, err := r.queries.CreateTwoFactor(ctx, generated.CreateTwoFactorParams{
		UserID:              userID,
		TotpSecretEncrypted: encryptedSecret,
	})
	if err != nil {
		return generated.UserTwoFactor{}, pgmap.TranslatePgErr(err)
	}

	return twoFactor, nil
}

func (r *TwoFactor) Upsert(ctx context.Context, userID pgtype.UUID, encryptedSecret string) (generated.UserTwoFactor, error) {
	twoFactor, err := r.queries.UpsertTwoFactor(ctx, generated.UpsertTwoFactorParams{
		UserID:              userID,
		TotpSecretEncrypted: encryptedSecret,
	})
	if err != nil {
		return generated.UserTwoFactor{}, pgmap.TranslatePgErr(err)
	}

	return twoFactor, nil
}

func (r *TwoFactor) MarkVerified(ctx context.Context, userID pgtype.UUID) (generated.UserTwoFactor, error) {
	twoFactor, err := r.queries.MarkTwoFactorVerified(ctx, userID)
	if err != nil {
		return generated.UserTwoFactor{}, pgmap.TranslatePgErr(err)
	}

	return twoFactor, nil
}

func (r *TwoFactor) Delete(ctx context.Context, userID pgtype.UUID) error {
	return pgmap.TranslatePgErr(r.queries.DeleteTwoFactor(ctx, userID))
}

func (r *TwoFactor) SetLastUsedCode(ctx context.Context, userID pgtype.UUID, code string) error {
	return pgmap.TranslatePgErr(r.queries.SetLastUsedCode(ctx, generated.SetLastUsedCodeParams{
		UserID: userID,
		Code:   pgtype.Text{String: code, Valid: true},
	}))
}
