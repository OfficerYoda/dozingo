package handler

import (
	"errors"
	"log/slog"

	"github.com/danielgtaylor/huma/v2"
	"github.com/jackc/pgx/v5"
)

// notFoundOr500 translates a database error into the appropriate huma error.
// pgx.ErrNoRows becomes a 404 with notFoundMsg; anything else is logged via
// slog and returned as a 500 with opMsg. opMsg is used both for the slog
// message and the client-facing error summary so server logs and API errors
// stay in lockstep.
func notFoundOr500(err error, notFoundMsg, opMsg string) error {
	if errors.Is(err, pgx.ErrNoRows) {
		return huma.Error404NotFound(notFoundMsg)
	}
	slog.Error(opMsg, "error", err)
	return huma.Error500InternalServerError(opMsg)
}

// internalError logs and returns a 500 for non-NotFound DB errors. Use when
// the operation can never legitimately produce ErrNoRows (e.g. list queries,
// bulk inserts) and a 404 would be incorrect.
func internalError(err error, opMsg string) error {
	slog.Error(opMsg, "error", err)
	return huma.Error500InternalServerError(opMsg)
}

// mapSlice returns a new slice with fn applied to every element of in.
// Always returns a non-nil slice (matching the prior "make([]T, 0)" pattern
// so JSON marshalling produces [] rather than null for empty results).
func mapSlice[T, U any](in []T, fn func(T) U) []U {
	out := make([]U, 0, len(in))
	for _, v := range in {
		out = append(out, fn(v))
	}
	return out
}
