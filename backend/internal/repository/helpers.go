package repository

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/officeryoda/dozingo/internal/domain"
)

func pgTextFromString(str *string) pgtype.Text {
	if str != nil && strings.TrimSpace(*str) != "" {
		return pgtype.Text{String: *str, Valid: true}
	}
	return pgtype.Text{Valid: false}
}

func stringFromPgText(text pgtype.Text) *string {
	if !text.Valid {
		return nil
	}
	s := text.String
	return &s
}

func stringFromPgUUID(uuid pgtype.UUID) *string {
	if !uuid.Valid {
		return nil
	}
	s := uuid.String()
	return &s
}

func pgTimestamptzFromTime(time time.Time) pgtype.Timestamptz {
	return pgtype.Timestamptz{
		Time:             time,
		InfinityModifier: pgtype.Finite,
		Valid:            true,
	}
}

func pgInt4FromInt32(v *int32) pgtype.Int4 {
	if v == nil {
		return pgtype.Int4{Valid: false}
	}
	return pgtype.Int4{Int32: *v, Valid: true}
}

// translatePgErr maps pgx/pgconn errors to domain sentinels. Anything we don't
// recognize is returned unchanged so it surfaces as a 500 in the handler.
func translatePgErr(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, pgx.ErrNoRows) {
		return fmt.Errorf("%w", domain.ErrNotFound)
	}
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case "23505": // unique_violation
			return fmt.Errorf("%s: %w", pgErr.ConstraintName, domain.ErrConflict)
		case "23503": // foreign_key_violation
			return fmt.Errorf("%s: %w", pgErr.ConstraintName, domain.ErrInvalid)
		}
	}
	return err
}
