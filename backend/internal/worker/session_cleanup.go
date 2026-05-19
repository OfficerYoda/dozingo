package worker

import (
	"context"
	"log/slog"
	"time"

	"github.com/officeryoda/dozingo/internal/generated"
)

type SessionCleaner struct {
	queries  *generated.Queries
	interval time.Duration
}

func NewSessionCleaner(queries *generated.Queries, interval time.Duration) *SessionCleaner {
	return &SessionCleaner{queries: queries, interval: interval}
}

func (s *SessionCleaner) Start(ctx context.Context) {
	go func() {
		// Run once at startup
		deleteExpiredSessions(ctx, s.queries)

		ticker := time.NewTicker(s.interval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				deleteExpiredSessions(ctx, s.queries)
			case <-ctx.Done():
				return
			}
		}
	}()
}

func deleteExpiredSessions(ctx context.Context, queries *generated.Queries) {
	err := queries.DeleteExpiredSessions(ctx)
	if err != nil {
		slog.Error("failed to clean up expired sessions", "error", err)
	}
}
