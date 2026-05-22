package pgmap

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

func PgTextFromString(str *string) pgtype.Text {
	if str != nil && strings.TrimSpace(*str) != "" {
		return pgtype.Text{String: *str, Valid: true}
	}
	return pgtype.Text{Valid: false}
}

func StringFromPgText(text pgtype.Text) *string {
	if !text.Valid {
		return nil
	}
	s := text.String
	return &s
}

func StringFromPgUUID(uuid pgtype.UUID) *string {
	if !uuid.Valid {
		return nil
	}
	s := uuid.String()
	return &s
}

func PgTimestamptzFromTime(time time.Time) pgtype.Timestamptz {
	return pgtype.Timestamptz{
		Time:             time,
		InfinityModifier: pgtype.Finite,
		Valid:            true,
	}
}

func PgInt4FromInt32(v *int32) pgtype.Int4 {
	if v == nil {
		return pgtype.Int4{Valid: false}
	}
	return pgtype.Int4{Int32: *v, Valid: true}
}

// TranslatePgErr maps pgx/pgconn errors to domain sentinels. Anything we don't
// recognize is returned unchanged so it surfaces as a 500 in the handler.
func TranslatePgErr(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.ErrNotFound
	}
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case "23505": // unique_violation
			return fmt.Errorf("%s: %w", pgErr.ConstraintName, domain.ErrConflict)
		case "23503": // foreign_key_violation
			return fmt.Errorf("%s: %w", pgErr.ConstraintName, domain.ErrBadInput)
		case "23514": // violates check constraint
			return fmt.Errorf("%s: %w", pgErr.ConstraintName, domain.ErrBadInput)
		}
	}
	return err
}
