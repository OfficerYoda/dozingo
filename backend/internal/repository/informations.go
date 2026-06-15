package repository

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/officeryoda/dozingo/internal/generated"
	"github.com/officeryoda/dozingo/internal/pgmap"
)

type Informations struct {
	queries *generated.Queries
}

func (r *Informations) GetSecurityInformation(ctx context.Context, userID pgtype.UUID) (generated.GetSecurityInformationRow, error) {
	infos, err := r.queries.GetSecurityInformation(ctx, userID)
	if err != nil {
		return generated.GetSecurityInformationRow{}, pgmap.TranslatePgErr(err)
	}

	return infos, nil
}
