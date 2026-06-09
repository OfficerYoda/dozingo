package service

import (
	"context"
	"fmt"
	"time"

	"github.com/officeryoda/dozingo/internal/generated"
	"github.com/officeryoda/dozingo/internal/pgmap"
	"github.com/officeryoda/dozingo/internal/repository"
)

type Stats struct {
	stats   *repository.Stats
	queries *generated.Queries
}

func NewStats(stats *repository.Stats, queries *generated.Queries) *Stats {
	return &Stats{
		stats:   stats,
		queries: queries,
	}
}

func (s *Stats) GetRecentStats(ctx context.Context, period time.Duration) (generated.GetRecentStatsRow, error) {
	stats, err := s.stats.GetRecentStats(ctx, pgmap.PgIntervalFromDuration(&period))
	if err != nil {
		return generated.GetRecentStatsRow{}, fmt.Errorf("get recent stats: %w", err)
	}

	return stats, nil
}
