package handler

import (
	"testing"

	"github.com/jackc/pgx/v5/pgtype"
)

func TestPgTextFromString_Nil(t *testing.T) {
	got := pgTextFromString(nil)
	if got.Valid {
		t.Errorf("nil input should yield invalid pgtype.Text, got %+v", got)
	}
}

func TestPgTextFromString_Empty(t *testing.T) {
	s := ""
	got := pgTextFromString(&s)
	if got.Valid {
		t.Errorf("empty string should yield invalid pgtype.Text, got %+v", got)
	}
}

func TestPgTextFromString_Whitespace(t *testing.T) {
	s := "   \t\n"
	got := pgTextFromString(&s)
	if got.Valid {
		t.Errorf("whitespace-only string should yield invalid pgtype.Text, got %+v", got)
	}
}

func TestPgTextFromString_NonEmpty(t *testing.T) {
	s := "hello"
	got := pgTextFromString(&s)
	if !got.Valid || got.String != "hello" {
		t.Errorf("expected {Valid:true String:%q}, got %+v", "hello", got)
	}
}

func TestStringFromPgText_Valid(t *testing.T) {
	in := pgtype.Text{String: "abc", Valid: true}
	got := stringFromPgText(in)
	if got == nil {
		t.Fatal("expected non-nil pointer")
	}
	if *got != "abc" {
		t.Errorf("expected %q, got %q", "abc", *got)
	}
}

func TestStringFromPgText_Invalid(t *testing.T) {
	in := pgtype.Text{String: "ignored", Valid: false}
	if got := stringFromPgText(in); got != nil {
		t.Errorf("expected nil pointer, got %q", *got)
	}
}

func TestStringFromPgUUID_Valid(t *testing.T) {
	var u pgtype.UUID
	if err := u.Scan("11111111-2222-3333-4444-555555555555"); err != nil {
		t.Fatalf("setup: failed to scan UUID: %v", err)
	}
	got := stringFromPgUUID(u)
	if got == nil {
		t.Fatal("expected non-nil pointer")
	}
	if *got != "11111111-2222-3333-4444-555555555555" {
		t.Errorf("unexpected UUID string: %q", *got)
	}
}

func TestStringFromPgUUID_Invalid(t *testing.T) {
	in := pgtype.UUID{Valid: false}
	if got := stringFromPgUUID(in); got != nil {
		t.Errorf("expected nil pointer, got %q", *got)
	}
}

func TestUuidFromString_Valid(t *testing.T) {
	got, err := uuidFromString("11111111-2222-3333-4444-555555555555")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !got.Valid {
		t.Fatal("expected Valid:true")
	}
}

func TestUuidFromString_Invalid(t *testing.T) {
	if _, err := uuidFromString("not-a-uuid"); err == nil {
		t.Fatal("expected error for non-UUID input")
	}
}

func TestUuidFromString_Empty(t *testing.T) {
	if _, err := uuidFromString(""); err == nil {
		t.Fatal("expected error for empty input")
	}
}
