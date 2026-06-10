// Package pgmap converts between pgx/pgtype values and Go types and translates Postgres errors to domain sentinels.
package pgmap

import (
	"errors"
	"fmt"
	"strings"
	timepkg "time"

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

func PgUUIDFromString(s *string) pgtype.UUID {
	if s == nil {
		return pgtype.UUID{Valid: false}
	}
	var uuid pgtype.UUID
	err := uuid.Scan(*s)
	if err != nil {
		return pgtype.UUID{Valid: false}
	}

	return uuid
}

func PgTimestamptzFromTime(time *timepkg.Time) pgtype.Timestamptz {
	if time == nil {
		return pgtype.Timestamptz{
			Time:             timepkg.Time{},
			InfinityModifier: pgtype.Finite,
			Valid:            false,
		}
	}

	return pgtype.Timestamptz{
		Time:             *time,
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

func PgIntervalFromDuration(d *timepkg.Duration) pgtype.Interval {
	if d == nil {
		return pgtype.Interval{Valid: false}
	}

	return pgtype.Interval{
		Microseconds: int64(*d / timepkg.Microsecond),
		Days:         0,
		Months:       0,
		Valid:        true,
	}
}

func DurationFromPgInterval(interval pgtype.Interval) *timepkg.Duration {
	if !interval.Valid {
		return nil
	}

	const (
		microsecond = int64(timepkg.Microsecond)
		hour        = int64(timepkg.Hour)
	)

	d := timepkg.Duration(
		interval.Microseconds*microsecond +
			int64(interval.Days)*24*hour +
			int64(interval.Months)*30*24*hour,
	)

	return &d
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
