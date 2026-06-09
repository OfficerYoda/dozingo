package pgmap

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/officeryoda/dozingo/internal/domain"
)

// ── PgTextFromString ──────────────────────────────────────────────────────────

func TestPgTextFromString_NonEmpty(t *testing.T) {
	s := "hello"
	got := PgTextFromString(&s)
	if !got.Valid {
		t.Fatal("expected Valid=true for a non-empty string")
	}
	if got.String != s {
		t.Errorf("expected String=%q, got %q", s, got.String)
	}
}

func TestPgTextFromString_Nil(t *testing.T) {
	got := PgTextFromString(nil)
	if got.Valid {
		t.Fatal("expected Valid=false for a nil pointer")
	}
}

func TestPgTextFromString_EmptyString(t *testing.T) {
	s := ""
	got := PgTextFromString(&s)
	if got.Valid {
		t.Fatal("expected Valid=false for an empty string")
	}
}

func TestPgTextFromString_WhitespaceOnly(t *testing.T) {
	s := "   \t\n"
	got := PgTextFromString(&s)
	if got.Valid {
		t.Fatal("expected Valid=false for a whitespace-only string")
	}
}

// ── StringFromPgText ──────────────────────────────────────────────────────────

func TestStringFromPgText_Valid(t *testing.T) {
	text := pgtype.Text{String: "world", Valid: true}
	got := StringFromPgText(text)
	if got == nil {
		t.Fatal("expected non-nil pointer for a valid pgtype.Text")
	}
	if *got != "world" {
		t.Errorf("expected %q, got %q", "world", *got)
	}
}

func TestStringFromPgText_Invalid(t *testing.T) {
	text := pgtype.Text{String: "ignored", Valid: false}
	got := StringFromPgText(text)
	if got != nil {
		t.Fatalf("expected nil for an invalid pgtype.Text, got %q", *got)
	}
}

func TestStringFromPgText_ReturnsDistinctPointer(t *testing.T) {
	text := pgtype.Text{String: "ptr", Valid: true}
	a := StringFromPgText(text)
	b := StringFromPgText(text)
	if a == b {
		t.Fatal("expected StringFromPgText to return a fresh pointer each call")
	}
}

// ── StringFromPgUUID ──────────────────────────────────────────────────────────

func TestStringFromPgUUID_Valid(t *testing.T) {
	const raw = "da78a8e3-506f-4e79-a50c-75c2b95156cc"
	var uuid pgtype.UUID
	if err := uuid.Scan(raw); err != nil {
		t.Fatalf("uuid.Scan failed: %v", err)
	}

	got := StringFromPgUUID(uuid)
	if got == nil {
		t.Fatal("expected non-nil pointer for a valid pgtype.UUID")
	}
	if *got != raw {
		t.Errorf("expected %q, got %q", raw, *got)
	}
}

func TestStringFromPgUUID_Invalid(t *testing.T) {
	var uuid pgtype.UUID // zero value — Valid is false
	got := StringFromPgUUID(uuid)
	if got != nil {
		t.Fatalf("expected nil for an invalid pgtype.UUID, got %q", *got)
	}
}

// ── PgUUIDFromString ──────────────────────────────────────────────────────────

func TestPgUUIDFromString_Valid(t *testing.T) {
	const raw = "da78a8e3-506f-4e79-a50c-75c2b95156cc"
	s := raw
	got := PgUUIDFromString(&s)
	if !got.Valid {
		t.Fatal("expected Valid=true for a well-formed UUID string")
	}

	// Round-trip through StringFromPgUUID to assert the bytes match the
	// canonical input. Comparing strings dodges the awkward [16]byte
	// formatting and matches how callers consume the value.
	round := StringFromPgUUID(got)
	if round == nil {
		t.Fatal("expected round-trip pointer to be non-nil")
	}
	if *round != raw {
		t.Errorf("expected round-trip %q, got %q", raw, *round)
	}
}

func TestPgUUIDFromString_Nil(t *testing.T) {
	got := PgUUIDFromString(nil)
	if got.Valid {
		t.Fatal("expected Valid=false for a nil pointer")
	}
}

func TestPgUUIDFromString_Invalid(t *testing.T) {
	s := "not-a-uuid"
	got := PgUUIDFromString(&s)
	if got.Valid {
		t.Fatal("expected Valid=false for an unparseable UUID string")
	}
}

func TestPgUUIDFromString_EmptyString(t *testing.T) {
	s := ""
	got := PgUUIDFromString(&s)
	if got.Valid {
		t.Fatal("expected Valid=false for an empty string")
	}
}

// ── PgTimestamptzFromTime ─────────────────────────────────────────────────────

func TestPgTimestamptzFromTime_Valid(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Microsecond)
	got := PgTimestamptzFromTime(&now)

	if !got.Valid {
		t.Fatal("expected Valid=true")
	}
	if got.InfinityModifier != pgtype.Finite {
		t.Errorf("expected InfinityModifier=Finite, got %v", got.InfinityModifier)
	}
	if !got.Time.Equal(now) {
		t.Errorf("expected Time=%v, got %v", now, got.Time)
	}
}

func TestPgTimestamptzFromTime_ZeroTime(t *testing.T) {
	got := PgTimestamptzFromTime(&time.Time{})
	if !got.Valid {
		t.Fatal("expected Valid=true even for the zero time")
	}
	if got.InfinityModifier != pgtype.Finite {
		t.Errorf("expected InfinityModifier=Finite, got %v", got.InfinityModifier)
	}
}

// ── PgInt4FromInt32 ───────────────────────────────────────────────────────────

func TestPgInt4FromInt32_Nil(t *testing.T) {
	got := PgInt4FromInt32(nil)
	if got.Valid {
		t.Fatal("expected Valid=false for a nil pointer")
	}
}

func TestPgInt4FromInt32_NonNil(t *testing.T) {
	v := int32(42)
	got := PgInt4FromInt32(&v)
	if !got.Valid {
		t.Fatal("expected Valid=true for a non-nil pointer")
	}
	if got.Int32 != v {
		t.Errorf("expected Int32=%d, got %d", v, got.Int32)
	}
}

func TestPgInt4FromInt32_Zero(t *testing.T) {
	v := int32(0)
	got := PgInt4FromInt32(&v)
	if !got.Valid {
		t.Fatal("expected Valid=true for a pointer to zero")
	}
	if got.Int32 != 0 {
		t.Errorf("expected Int32=0, got %d", got.Int32)
	}
}

func TestPgInt4FromInt32_Negative(t *testing.T) {
	v := int32(-1)
	got := PgInt4FromInt32(&v)
	if !got.Valid {
		t.Fatal("expected Valid=true for a negative value")
	}
	if got.Int32 != v {
		t.Errorf("expected Int32=%d, got %d", v, got.Int32)
	}
}

// ── TranslatePgErr ────────────────────────────────────────────────────────────

func TestTranslatePgErr_Nil(t *testing.T) {
	if got := TranslatePgErr(nil); got != nil {
		t.Fatalf("expected nil for a nil error, got %v", got)
	}
}

func TestTranslatePgErr_ErrNoRows(t *testing.T) {
	got := TranslatePgErr(pgx.ErrNoRows)
	if !errors.Is(got, domain.ErrNotFound) {
		t.Errorf("expected domain.ErrNotFound, got %v", got)
	}
}

func TestTranslatePgErr_UniqueViolation(t *testing.T) {
	pgErr := &pgconn.PgError{
		Code:           "23505",
		ConstraintName: "users_email_key",
	}
	got := TranslatePgErr(pgErr)
	if !errors.Is(got, domain.ErrConflict) {
		t.Errorf("expected domain.ErrConflict, got %v", got)
	}
	if got.Error() == "" {
		t.Fatal("expected a non-empty error message")
	}
	// Constraint name must be part of the message for easy debugging.
	want := "users_email_key"
	if !containsString(got.Error(), want) {
		t.Errorf("expected constraint name %q in error message %q", want, got.Error())
	}
}

func TestTranslatePgErr_ForeignKeyViolation(t *testing.T) {
	pgErr := &pgconn.PgError{
		Code:           "23503",
		ConstraintName: "votes_user_id_fkey",
	}
	got := TranslatePgErr(pgErr)
	if !errors.Is(got, domain.ErrBadInput) {
		t.Errorf("expected domain.ErrInvalid, got %v", got)
	}
	want := "votes_user_id_fkey"
	if !containsString(got.Error(), want) {
		t.Errorf("expected constraint name %q in error message %q", want, got.Error())
	}
}

func TestTranslatePgErr_UnrecognizedPgCode(t *testing.T) {
	pgErr := &pgconn.PgError{Code: "42P01"} // undefined_table
	got := TranslatePgErr(pgErr)
	// Must be passed through unchanged so it bubbles up as a 500.
	if !errors.Is(got, pgErr) {
		t.Errorf("expected the original pgError to be returned unchanged, got %v", got)
	}
}

func TestTranslatePgErr_WrappedErrNoRows(t *testing.T) {
	wrapped := fmt.Errorf("repo: %w", pgx.ErrNoRows)
	got := TranslatePgErr(wrapped)
	if !errors.Is(got, domain.ErrNotFound) {
		t.Errorf("expected domain.ErrNotFound for a wrapped ErrNoRows, got %v", got)
	}
}

func TestTranslatePgErr_GenericError(t *testing.T) {
	sentinel := errors.New("some unexpected db error")
	got := TranslatePgErr(sentinel)
	if !errors.Is(got, sentinel) {
		t.Errorf("expected the original error to be returned unchanged, got %v", got)
	}
}

// ── helpers ───────────────────────────────────────────────────────────────────

func containsString(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr ||
		len(substr) == 0 ||
		func() bool {
			for i := 0; i <= len(s)-len(substr); i++ {
				if s[i:i+len(substr)] == substr {
					return true
				}
			}
			return false
		}())
}
