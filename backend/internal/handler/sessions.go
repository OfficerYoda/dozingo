package handler

import (
	"context"
	"log/slog"
	"time"

	"github.com/officeryoda/dozingo/internal/generated"
)

func StartSessionCleanup(ctx context.Context, queries *generated.Queries) {
	go func() {
		// Run once at startup
		deleteExpiredSessions(ctx, queries)

		ticker := time.NewTicker(24 * time.Hour)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				deleteExpiredSessions(ctx, queries)
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
