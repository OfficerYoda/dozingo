package handler

import (
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
	if s != nil {
		return pgtype.Text{String: *s, Valid: true}
	} else {
		return pgtype.Text{String: "", Valid: true}
	}
}

func stringFromPgText(pgText pgtype.Text) *string {
	if pgText.Valid {
		s := pgText.String
		return &s
	}
	return nil
}

func stringFromPgUUID(pgText pgtype.UUID) *string {
	if pgText.Valid {
		s := pgText.String()
		return &s
	}
	return nil
}
