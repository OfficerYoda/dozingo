package repository

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/officeryoda/dozingo/internal/generated"
	"github.com/officeryoda/dozingo/internal/pgmap"
)

type RecoveryCodes struct {
	queries *generated.Queries
}

func (r *RecoveryCodes) Create(ctx context.Context, userID pgtype.UUID, codeHashes []string) ([]generated.RecoveryCode, error) {
	code, err := r.queries.CreateRecoveryCodes(ctx, generated.CreateRecoveryCodesParams{
		UserID:     userID,
		CodeHashes: codeHashes,
	})
	if err != nil {
		return []generated.RecoveryCode{}, pgmap.TranslatePgErr(err)
	}

	return code, nil
}

func (r *RecoveryCodes) GetUnusedByUserID(ctx context.Context, userID pgtype.UUID) ([]generated.RecoveryCode, error) {
	codes, err := r.queries.GetUnusedRecoveryCodesByUserID(ctx, userID)
	if err != nil {
		return nil, pgmap.TranslatePgErr(err)
	}

	return codes, nil
}

func (r *RecoveryCodes) MarkUsed(ctx context.Context, codeID pgtype.UUID) (generated.RecoveryCode, error) {
	code, err := r.queries.MarkRecoveryCodeUsed(ctx, codeID)
	if err != nil {
		return generated.RecoveryCode{}, pgmap.TranslatePgErr(err)
	}

	return code, nil
}

func (r *RecoveryCodes) CountUnusedByUserID(ctx context.Context, userID pgtype.UUID) (int64, error) {
	count, err := r.queries.CountUnusedRecoveryCodesByUserID(ctx, userID)
	if err != nil {
		return 0, pgmap.TranslatePgErr(err)
	}

	return count, nil
}

func (r *RecoveryCodes) DeleteByUserID(ctx context.Context, userID pgtype.UUID) error {
	return pgmap.TranslatePgErr(r.queries.DeleteRecoveryCodesByUserID(ctx, userID))
}
