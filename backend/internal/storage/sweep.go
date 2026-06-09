package storage

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	s3types "github.com/aws/aws-sdk-go-v2/service/s3/types"
)

// DefaultAvatarKey shuold not be deleted on sweep
const DefaultAvatarKey = "default.svg"

type InUseKeyLister interface {
	ListInUseAvatarKeys(ctx context.Context) ([]string, error)
}

type SweepConfig struct {
	// GracePeriod is the minimum age before deletion is allowed
	GracePeriod  time.Duration
	MaxDeletions int
	// Dry, when true, runs a dry run
	Dry bool
}

func DefaultSweepConfig() SweepConfig {
	return SweepConfig{
		GracePeriod:  5 * time.Minute,
		MaxDeletions: 1000,
		Dry:          false,
	}
}

func (g *Garage) SweepOrphanAvatars(ctx context.Context, lister InUseKeyLister, cfg SweepConfig) error {
	start := time.Now()

	inUse, err := g.buildProtectedKeySet(ctx, lister)
	if err != nil {
		return err
	}
	if inUse == nil {
		return nil
	}

	collected, err := g.collectCandidates(ctx, inUse, cfg)
	if err != nil {
		return err
	}
	if collected == nil {
		return nil
	}

	missingInBucket := g.warnMissingKeys(inUse, collected.seenInBucket)

	deleted, err := g.deleteCandidates(ctx, collected.candidates, cfg)
	if err != nil {
		return err
	}

	g.logSweepResult(sweepMetrics{
		scanned:           len(collected.seenInBucket),
		inUseCount:        len(inUse),
		deleted:           deleted,
		candidates:        len(collected.candidates),
		skippedDueToGrace: collected.skippedDueToGrace,
		missingInBucket:   missingInBucket,
		hardCapReached:    collected.hardCapReached,
		dry:               cfg.Dry,
		duration:          time.Since(start),
	})

	return nil
}

func (g *Garage) buildProtectedKeySet(ctx context.Context, lister InUseKeyLister) (map[string]struct{}, error) {
	dbKeys, err := lister.ListInUseAvatarKeys(ctx)
	if err != nil {
		return nil, fmt.Errorf("list in-use avatar keys: %w", err)
	}

	if len(dbKeys) == 0 {
		slog.Warn("avatar_orphan_cleanup skipped: no in-use keys returned from DB",
			"bucket", g.bucketName)
		return nil, nil
	}

	inUse := make(map[string]struct{}, len(dbKeys)+1)
	for _, k := range dbKeys {
		inUse[k] = struct{}{}
	}
	inUse[DefaultAvatarKey] = struct{}{}

	return inUse, nil
}

type collectResult struct {
	candidates        []s3types.ObjectIdentifier
	seenInBucket      map[string]struct{}
	skippedDueToGrace int
	hardCapReached    bool
}

func (g *Garage) collectCandidates(ctx context.Context, inUse map[string]struct{}, cfg SweepConfig) (*collectResult, error) {
	var res collectResult
	cutoff := time.Now().Add(-cfg.GracePeriod)
	res.seenInBucket = make(map[string]struct{})
	var continuationToken *string

paging:
	for {
		if err := ctx.Err(); err != nil {
			return nil, err
		}

		out, listErr := g.s3Client.ListObjectsV2(ctx, &s3.ListObjectsV2Input{
			Bucket:            aws.String(g.bucketName),
			ContinuationToken: continuationToken,
		})
		if listErr != nil {
			if errors.As(listErr, new(*s3types.NoSuchBucket)) {
				slog.Warn("avatar_orphan_cleanup: bucket does not exist", "bucket", g.bucketName)
				return nil, nil
			}
			return nil, fmt.Errorf("list objects: %w", listErr)
		}

		for _, obj := range out.Contents {
			if obj.Key == nil {
				continue
			}
			key := *obj.Key
			res.seenInBucket[key] = struct{}{}

			if g.shouldSkip(key, obj.LastModified, cutoff, inUse) {
				if obj.LastModified != nil && obj.LastModified.After(cutoff) {
					res.skippedDueToGrace++
				}
				continue
			}

			res.candidates = append(res.candidates, s3types.ObjectIdentifier{Key: aws.String(key)})

			if len(res.candidates) >= cfg.MaxDeletions {
				res.hardCapReached = true
				break paging
			}
		}

		if out.IsTruncated == nil || !*out.IsTruncated {
			break
		}
		continuationToken = out.NextContinuationToken
	}

	return &res, nil
}

func (g *Garage) shouldSkip(key string, lastModified *time.Time, cutoff time.Time, inUse map[string]struct{}) bool {
	if key == DefaultAvatarKey {
		return true
	}
	if _, ok := inUse[key]; ok {
		return true
	}
	if lastModified != nil && lastModified.After(cutoff) {
		return true
	}
	return false
}

func (g *Garage) warnMissingKeys(inUse, seenInBucket map[string]struct{}) int {
	missing := 0
	for k := range inUse {
		if k == DefaultAvatarKey {
			continue
		}
		if _, ok := seenInBucket[k]; !ok {
			slog.Warn("avatar referenced by DB but missing in bucket",
				"key", k, "bucket", g.bucketName)
			missing++
		}
	}
	return missing
}

func (g *Garage) deleteCandidates(ctx context.Context, candidates []s3types.ObjectIdentifier, cfg SweepConfig) (int, error) {
	if len(candidates) == 0 || cfg.Dry {
		return 0, nil
	}

	_, err := g.s3Client.DeleteObjects(ctx, &s3.DeleteObjectsInput{
		Bucket: aws.String(g.bucketName),
		Delete: &s3types.Delete{
			Objects: candidates,
			Quiet:   aws.Bool(true),
		},
	})
	if err != nil {
		return 0, fmt.Errorf("delete objects: %w", err)
	}

	return len(candidates), nil
}

type sweepMetrics struct {
	scanned           int
	inUseCount        int
	deleted           int
	candidates        int
	skippedDueToGrace int
	missingInBucket   int
	hardCapReached    bool
	dry               bool
	duration          time.Duration
}

func (g *Garage) logSweepResult(m sweepMetrics) {
	slog.Info(
		"avatar_orphan_cleanup",
		"bucket", g.bucketName,
		"scanned", m.scanned,
		"in_use_count", m.inUseCount,
		"deleted", m.deleted,
		"candidates", m.candidates,
		"skipped_due_to_grace", m.skippedDueToGrace,
		"missing_in_bucket", m.missingInBucket,
		"hard_cap_reached", m.hardCapReached,
		"dry_run", m.dry,
		"duration", m.duration,
	)
}
