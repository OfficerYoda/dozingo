package repository

import (
	"strings"

	"github.com/jackc/pgx/v5/pgtype"
)

// pgTextFromString converts a *string into pgtype.Text.
func pgTextFromString(s *string) pgtype.Text {
	if s != nil && strings.TrimSpace(*s) != "" {
		return pgtype.Text{String: *s, Valid: true}
	}
	return pgtype.Text{Valid: false}
}

// stringFromPgText converts a pgtype.Text to *string, returning nil on NULL.
func stringFromPgText(t pgtype.Text) *string {
	if !t.Valid {
		return nil
	}
	s := t.String
	return &s
}

// stringFromPgUUID converts a pgtype.UUID to *string, returning nil on NULL.
func stringFromPgUUID(u pgtype.UUID) *string {
	if !u.Valid {
		return nil
	}
	s := u.String()
	return &s
}
