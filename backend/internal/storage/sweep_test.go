package storage

import (
	"context"
	"errors"
	"sort"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	s3types "github.com/aws/aws-sdk-go-v2/service/s3/types"
)

/// ===== Fakes =====

// fakeS3 is an in-memory s3API used by sweep tests. It supports
// ListObjectsV2 (single-page or multi-page via continuation tokens),
// DeleteObjects, and injectable error modes.
type fakeS3 struct {
	bucketName string
	objects    map[string]time.Time // key -> LastModified

	// pageSize controls how many objects ListObjectsV2 returns per page.
	// 0 means "everything in one page".
	pageSize int

	// listErr (if non-nil) is returned from ListObjectsV2.
	listErr error
	// deleteErr (if non-nil) is returned from DeleteObjects.
	deleteErr error

	// deletes records every batch passed to DeleteObjects, in call order.
	deletes [][]string
}

func (f *fakeS3) PutObject(ctx context.Context, in *s3.PutObjectInput, opts ...func(*s3.Options)) (*s3.PutObjectOutput, error) {
	// Not used by sweep tests; left as a stub so the type satisfies s3API.
	return &s3.PutObjectOutput{}, nil
}

func (f *fakeS3) ListObjectsV2(ctx context.Context, in *s3.ListObjectsV2Input, opts ...func(*s3.Options)) (*s3.ListObjectsV2Output, error) {
	if f.listErr != nil {
		return nil, f.listErr
	}

	// Collect keys in deterministic order so paging is reproducible.
	keys := make([]string, 0, len(f.objects))
	for k := range f.objects {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// Resolve continuation: token "after-K" means "first key strictly
	// greater than K". Empty token means "from start".
	startIdx := 0
	if in.ContinuationToken != nil {
		ct := *in.ContinuationToken
		for i, k := range keys {
			if k > ct {
				startIdx = i
				break
			}
			startIdx = i + 1
		}
	}

	pageSize := f.pageSize
	if pageSize <= 0 {
		pageSize = len(keys)
	}

	end := startIdx + pageSize
	if end > len(keys) {
		end = len(keys)
	}

	page := keys[startIdx:end]
	contents := make([]s3types.Object, 0, len(page))
	for _, k := range page {
		k := k
		lm := f.objects[k]
		contents = append(contents, s3types.Object{
			Key:          aws.String(k),
			LastModified: &lm,
		})
	}

	out := &s3.ListObjectsV2Output{Contents: contents}
	if end < len(keys) {
		truncated := true
		out.IsTruncated = &truncated
		// Continuation token = last key returned. Next call resumes
		// from "first key > token".
		token := page[len(page)-1]
		out.NextContinuationToken = &token
	}
	return out, nil
}

func (f *fakeS3) DeleteObjects(ctx context.Context, in *s3.DeleteObjectsInput, opts ...func(*s3.Options)) (*s3.DeleteObjectsOutput, error) {
	if f.deleteErr != nil {
		return nil, f.deleteErr
	}
	if in == nil || in.Delete == nil {
		return &s3.DeleteObjectsOutput{}, nil
	}
	keys := make([]string, 0, len(in.Delete.Objects))
	for _, o := range in.Delete.Objects {
		if o.Key != nil {
			keys = append(keys, *o.Key)
			delete(f.objects, *o.Key)
		}
	}
	f.deletes = append(f.deletes, keys)
	return &s3.DeleteObjectsOutput{}, nil
}

// fakeLister is an in-memory InUseKeyLister.
type fakeLister struct {
	keys []string
	err  error
}

func (f *fakeLister) ListInUseAvatarKeys(ctx context.Context) ([]string, error) {
	if f.err != nil {
		return nil, f.err
	}
	// Return a copy so the sweep can't mutate test state through the slice.
	out := make([]string, len(f.keys))
	copy(out, f.keys)
	return out, nil
}

/// ===== Helpers =====

func newGarageWithFake(f *fakeS3) *Garage {
	return &Garage{
		bucketName: f.bucketName,
		s3Client:   f,
	}
}

func tightCfg() SweepConfig {
	// Zero grace makes time-based assertions deterministic; tests that
	// need grace-period behaviour override this explicitly.
	return SweepConfig{GracePeriod: 0, MaxDeletions: 1000}
}

// recordedDeletes returns every key passed to DeleteObjects across all
// batches, sorted for stable comparison.
func recordedDeletes(f *fakeS3) []string {
	var all []string
	for _, batch := range f.deletes {
		all = append(all, batch...)
	}
	sort.Strings(all)
	return all
}

/// ===== Tests =====

func TestSweep_EmptyBucket_NoDeletes(t *testing.T) {
	f := &fakeS3{bucketName: "pics", objects: map[string]time.Time{}}
	g := newGarageWithFake(f)

	err := g.SweepOrphanAvatars(context.Background(),
		&fakeLister{keys: []string{"a.svg"}}, tightCfg())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(f.deletes) != 0 {
		t.Errorf("expected no DeleteObjects calls, got %v", f.deletes)
	}
}

func TestSweep_OnlyDefaultAndInUse_NoDeletes(t *testing.T) {
	f := &fakeS3{
		bucketName: "pics",
		objects: map[string]time.Time{
			DefaultAvatarKey: time.Now().Add(-time.Hour),
			"alice.svg":      time.Now().Add(-time.Hour),
			"bob.svg":        time.Now().Add(-time.Hour),
		},
	}
	g := newGarageWithFake(f)

	err := g.SweepOrphanAvatars(context.Background(),
		&fakeLister{keys: []string{"alice.svg", "bob.svg"}}, tightCfg())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(f.deletes) != 0 {
		t.Errorf("expected no deletes, got %v", f.deletes)
	}
}

func TestSweep_DeletesOrphanOlderThanCutoff(t *testing.T) {
	f := &fakeS3{
		bucketName: "pics",
		objects: map[string]time.Time{
			"alice.svg":  time.Now().Add(-time.Hour),
			"orphan.svg": time.Now().Add(-time.Hour),
		},
	}
	g := newGarageWithFake(f)

	err := g.SweepOrphanAvatars(context.Background(),
		&fakeLister{keys: []string{"alice.svg"}}, tightCfg())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got := recordedDeletes(f)
	want := []string{"orphan.svg"}
	if !equalStrings(got, want) {
		t.Errorf("expected deletes %v, got %v", want, got)
	}
}

func TestSweep_FreshOrphanInsideGracePeriod_NotDeleted(t *testing.T) {
	now := time.Now()
	f := &fakeS3{
		bucketName: "pics",
		objects: map[string]time.Time{
			"alice.svg":  now.Add(-time.Hour),
			"orphan.svg": now.Add(-1 * time.Minute), // inside grace
		},
	}
	g := newGarageWithFake(f)

	cfg := SweepConfig{GracePeriod: 5 * time.Minute, MaxDeletions: 1000}
	err := g.SweepOrphanAvatars(context.Background(),
		&fakeLister{keys: []string{"alice.svg"}}, cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(f.deletes) != 0 {
		t.Errorf("expected fresh orphan to be skipped, got deletes %v", f.deletes)
	}
}

func TestSweep_DefaultIsProtectedEvenWhenUnreferenced(t *testing.T) {
	// default.svg exists in the bucket but no row references it.
	// Sweep must still leave it alone.
	f := &fakeS3{
		bucketName: "pics",
		objects: map[string]time.Time{
			DefaultAvatarKey: time.Now().Add(-time.Hour),
			"alice.svg":      time.Now().Add(-time.Hour),
		},
	}
	g := newGarageWithFake(f)

	err := g.SweepOrphanAvatars(context.Background(),
		&fakeLister{keys: []string{"alice.svg"}}, tightCfg())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(f.deletes) != 0 {
		t.Errorf("expected default.svg to be protected, got deletes %v", f.deletes)
	}
}

func TestSweep_InUseKeyMissingFromBucket_NotDeleted(t *testing.T) {
	// "alice.svg" is referenced in the DB but doesn't exist in the bucket.
	// The sweep must not delete anything (there's nothing to delete) and
	// must not error. (The slog.Warn it emits is observable but we don't
	// assert on it here.)
	f := &fakeS3{
		bucketName: "pics",
		objects: map[string]time.Time{
			"bob.svg": time.Now().Add(-time.Hour),
		},
	}
	g := newGarageWithFake(f)

	err := g.SweepOrphanAvatars(context.Background(),
		&fakeLister{keys: []string{"alice.svg", "bob.svg"}}, tightCfg())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(f.deletes) != 0 {
		t.Errorf("expected no deletes, got %v", f.deletes)
	}
}

func TestSweep_HardCap_DeletesUpToLimit(t *testing.T) {
	objects := map[string]time.Time{
		"alice.svg": time.Now().Add(-time.Hour),
	}
	// 1500 orphans; cap = 1000 -> exactly 1000 deleted, 500 survive.
	for i := 0; i < 1500; i++ {
		objects[fmtKey(i)] = time.Now().Add(-time.Hour)
	}
	f := &fakeS3{bucketName: "pics", objects: objects, pageSize: 200}
	g := newGarageWithFake(f)

	cfg := SweepConfig{GracePeriod: 0, MaxDeletions: 1000}
	err := g.SweepOrphanAvatars(context.Background(),
		&fakeLister{keys: []string{"alice.svg"}}, cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	deleted := 0
	for _, batch := range f.deletes {
		deleted += len(batch)
	}
	if deleted != 1000 {
		t.Errorf("expected 1000 deletions, got %d", deleted)
	}
	// Bucket should now have alice.svg + 500 surviving orphans.
	if got := len(f.objects); got != 501 {
		t.Errorf("expected 501 surviving objects, got %d", got)
	}
}

func TestSweep_NoSuchBucket_LoggedNoOp(t *testing.T) {
	nsb := &s3types.NoSuchBucket{}
	f := &fakeS3{
		bucketName: "missing",
		objects:    map[string]time.Time{},
		listErr:    nsb,
	}
	g := newGarageWithFake(f)

	err := g.SweepOrphanAvatars(context.Background(),
		&fakeLister{keys: []string{"alice.svg"}}, tightCfg())
	if err != nil {
		t.Errorf("expected NoSuchBucket to be a no-op, got error: %v", err)
	}
	if len(f.deletes) != 0 {
		t.Errorf("expected no deletes on missing bucket, got %v", f.deletes)
	}
}

func TestSweep_EmptyInUseSet_BailsWithoutDeleting(t *testing.T) {
	// Lister returns nothing (e.g. truly empty users table, or a DB
	// glitch). Sweep must not delete anything in the bucket -- otherwise
	// a transient DB issue could wipe everyone's avatars.
	f := &fakeS3{
		bucketName: "pics",
		objects: map[string]time.Time{
			"alice.svg":  time.Now().Add(-time.Hour),
			"orphan.svg": time.Now().Add(-time.Hour),
		},
	}
	g := newGarageWithFake(f)

	err := g.SweepOrphanAvatars(context.Background(),
		&fakeLister{keys: nil}, tightCfg())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(f.deletes) != 0 {
		t.Errorf("expected no deletes when in-use set is empty, got %v", f.deletes)
	}
}

func TestSweep_DeleteObjectsError_Propagates(t *testing.T) {
	f := &fakeS3{
		bucketName: "pics",
		objects: map[string]time.Time{
			"orphan.svg": time.Now().Add(-time.Hour),
		},
		deleteErr: errors.New("garage exploded"),
	}
	g := newGarageWithFake(f)

	err := g.SweepOrphanAvatars(context.Background(),
		&fakeLister{keys: []string{"alice.svg"}}, tightCfg())
	if err == nil {
		t.Fatal("expected error from DeleteObjects, got nil")
	}
}

func TestSweep_ListerError_Propagates(t *testing.T) {
	f := &fakeS3{bucketName: "pics", objects: map[string]time.Time{}}
	g := newGarageWithFake(f)

	err := g.SweepOrphanAvatars(context.Background(),
		&fakeLister{err: errors.New("db down")}, tightCfg())
	if err == nil {
		t.Fatal("expected error when lister fails, got nil")
	}
	if len(f.deletes) != 0 {
		t.Errorf("expected no deletes when lister fails, got %v", f.deletes)
	}
}

func TestSweep_PaginatesAcrossMultiplePages(t *testing.T) {
	objects := map[string]time.Time{
		"alice.svg": time.Now().Add(-time.Hour),
	}
	for i := 0; i < 250; i++ {
		objects[fmtKey(i)] = time.Now().Add(-time.Hour)
	}
	// pageSize 100 -> 3 pages (100, 100, 51).
	f := &fakeS3{bucketName: "pics", objects: objects, pageSize: 100}
	g := newGarageWithFake(f)

	err := g.SweepOrphanAvatars(context.Background(),
		&fakeLister{keys: []string{"alice.svg"}}, tightCfg())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	deleted := 0
	for _, batch := range f.deletes {
		deleted += len(batch)
	}
	// 250 orphans should all be deleted (cap == 1000).
	if deleted != 250 {
		t.Errorf("expected 250 deletions across paginated listing, got %d", deleted)
	}
	// Bucket should retain only alice.svg.
	if got := len(f.objects); got != 1 {
		t.Errorf("expected only alice.svg to survive, got %d objects", got)
	}
}

func TestSweep_DryRun_RecordsCandidatesButDoesNotDelete(t *testing.T) {
	f := &fakeS3{
		bucketName: "pics",
		objects: map[string]time.Time{
			"alice.svg":  time.Now().Add(-time.Hour),
			"orphan.svg": time.Now().Add(-time.Hour),
		},
	}
	g := newGarageWithFake(f)

	cfg := SweepConfig{GracePeriod: 0, MaxDeletions: 1000, Dry: true}
	err := g.SweepOrphanAvatars(context.Background(),
		&fakeLister{keys: []string{"alice.svg"}}, cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(f.deletes) != 0 {
		t.Errorf("dry-run must not call DeleteObjects, got %v", f.deletes)
	}
	// Bucket state unchanged.
	if _, ok := f.objects["orphan.svg"]; !ok {
		t.Error("dry-run must leave orphan.svg in place")
	}
}

/// ===== utility =====

func equalStrings(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// fmtKey makes deterministic, lexicographically ordered keys so paging
// in the fake s3 is predictable.
func fmtKey(i int) string {
	const digits = "0123456789"
	// 4-digit zero-padded "orphan-NNNN.svg"
	buf := []byte("orphan-0000.svg")
	for d, base := 0, 1000; base > 0; d, base = d+1, base/10 {
		buf[7+d] = digits[(i/base)%10]
	}
	return string(buf)
}
