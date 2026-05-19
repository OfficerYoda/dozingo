package handler

import (
	"strings"

	"github.com/jackc/pgx/v5/pgtype"
)

func uuidFromString(s string) (pgtype.UUID, error) {
	var id pgtype.UUID
	if err := id.Scan(s); err != nil {
		return pgtype.UUID{}, err
	}
	return id, nil
}

func pgTextFromString(s *string) pgtype.Text {
	if s != nil && strings.TrimSpace(*s) != "" {
		return pgtype.Text{String: *s, Valid: true}
	} else {
		return pgtype.Text{String: "", Valid: false}
	}
}

func stringFromPgText(pgText pgtype.Text) *string {
	if pgText.Valid {
		s := pgText.String
		return &s
	}
	return nil
}

func stringFromPgUUID(pgUUID pgtype.UUID) *string {
	if pgUUID.Valid {
		s := pgUUID.String()
		return &s
	}
	return nil
}
