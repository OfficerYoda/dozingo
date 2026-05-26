package worker

import (
	"context"
	"log/slog"
	"time"

	"github.com/officeryoda/dozingo/internal/repository"
)

type SessionCleaner struct {
	sessions *repository.Sessions
	interval time.Duration
}

func NewSessionCleaner(sessions *repository.Sessions, interval time.Duration) *SessionCleaner {
	return &SessionCleaner{sessions: sessions, interval: interval}
}

func (s *SessionCleaner) Start(ctx context.Context) {
	go func() {
		// Run once at startup
		deleteExpiredSessions(ctx, s.sessions)

		ticker := time.NewTicker(s.interval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				deleteExpiredSessions(ctx, s.sessions)
			case <-ctx.Done():
				return
			}
		}
	}()
}

func deleteExpiredSessions(ctx context.Context, sessions *repository.Sessions) {
	err := sessions.DeleteExpiredSessions(ctx)
	if err != nil {
		slog.Error("failed to clean up expired sessions", "error", err)
	}
}
