package repository

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/officeryoda/dozingo/internal/generated"
	"github.com/officeryoda/dozingo/internal/pgmap"
)

type Stats struct {
	queries *generated.Queries
}

func (r *Stats) GetRecentStats(ctx context.Context, period pgtype.Interval) (generated.GetRecentStatsRow, error) {
	stats, err := r.queries.GetRecentStats(ctx, period)
	if err != nil {
		return generated.GetRecentStatsRow{}, pgmap.TranslatePgErr(err)
	}

	return stats, nil
}
