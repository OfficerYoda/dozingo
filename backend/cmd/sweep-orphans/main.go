package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/officeryoda/dozingo/internal/config"
	"github.com/officeryoda/dozingo/internal/repository"
	"github.com/officeryoda/dozingo/internal/storage"
)

// sweep-orphans runs orphan-cleanup sweeps on demand.
//
// Currently a single sweep target is supported: avatar objects in the
// Garage bucket whose key isn't referenced by any users.avatar_key. The
// API runs the same sweep hourly as a periodic worker; this CLI exists so
// devs and ops can trigger it on demand, inspect a dry-run before the
// next tick, or run it from an external cron.
//
// Concurrency note: it is safe to run this concurrently with the API's
// periodic worker. S3 DeleteObjects is idempotent; in the worst case both
// callers attempt to delete the same key and one log line says "deleted"
// twice.
func main() {
	dryRun := flag.Bool("dry-run", false, "list candidate orphans without deleting them")
	grace := flag.Duration("grace", 5*time.Minute,
		"minimum age an object must have before it's eligible for deletion")
	maxDeletions := flag.Int("max", 1000,
		"hard cap on deletions per run (S3 DeleteObjects accepts up to 1000 keys)")
	flag.Parse()

	if err := run(*dryRun, *grace, *maxDeletions); err != nil {
		slog.Error("sweep-orphans failed", "error", err)
		os.Exit(1)
	}
}

func run(dryRun bool, grace time.Duration, maxDeletions int) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	ctx := context.Background()

	pool, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		return fmt.Errorf("create pool: %w", err)
	}
	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		return fmt.Errorf("ping database: %w", err)
	}

	repos := repository.New(pool)
	garage := storage.NewGarage(ctx, cfg)

	sweepCfg := storage.SweepConfig{
		GracePeriod:  grace,
		MaxDeletions: maxDeletions,
		Dry:          dryRun,
	}

	slog.Info(
		"sweep-orphans starting",
		"target", "avatars",
		"dry_run", sweepCfg.Dry,
		"grace_period", sweepCfg.GracePeriod,
		"max_deletions", sweepCfg.MaxDeletions,
	)

	if err := garage.SweepOrphanAvatars(ctx, repos.Users, sweepCfg); err != nil {
		return fmt.Errorf("sweep avatars: %w", err)
	}

	return nil
}
