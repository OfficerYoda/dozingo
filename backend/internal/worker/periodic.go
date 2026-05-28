package worker

import (
	"context"
	"log/slog"
	"time"
)

type Periodic struct {
	name     string
	interval time.Duration
	run      func(context.Context) error
}

func NewPeriodic(name string, interval time.Duration, run func(context.Context) error) *Periodic {
	return &Periodic{name: name, interval: interval, run: run}
}

// Start spawns a goroutine that executes the job once immediately, then on
// every tick of the configured interval, until ctx is cancelled.
func (p *Periodic) Start(ctx context.Context) {
	go func() {
		p.runOnce(ctx)

		ticker := time.NewTicker(p.interval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				p.runOnce(ctx)
			case <-ctx.Done():
				return
			}
		}
	}()
}

func (p *Periodic) runOnce(ctx context.Context) {
	if err := p.run(ctx); err != nil {
		slog.Error("periodic job failed", "job", p.name, "error", err)
	}
}
