package handler

import (
	"errors"
	"log/slog"

	"github.com/danielgtaylor/huma/v2"
	"github.com/officeryoda/dozingo/internal/domain"
)

const (
	msgNotFound            = "not found"
	msgConflict            = "conflict"
	msgUnauthorized        = "unauthorized"
	msgForbidden           = "forbidden"
	msgBadRequest          = "bad request"
	msgUnprocessableEntity = "unprocessable entity"
)

// toHumaErr maps a domain error to a huma HTTP error. Sentinel-matched
// errors get a fixed user-facing message; the wrapped error is logged so
// constraint names and other internal context remain available
// server-side. Unmatched errors surface as 500 with opMsg as the body.
//
// notFoundMsg overrides the default "not found" message for ErrNotFound
// when non-empty (e.g. "board not found"). It is the only per-call
// override: other sentinels deliberately use generic messages to avoid
// callers accidentally re-introducing the leak by passing err.Error().
func toHumaErr(err error, notFoundMsg, opMsg string) error {
	switch {
	case errors.Is(err, domain.ErrNotFound):
		slog.Warn(opMsg, "error", err)
		msg := notFoundMsg
		if msg == "" {
			msg = msgNotFound
		}
		return huma.Error404NotFound(msg)
	case errors.Is(err, domain.ErrConflict):
		slog.Warn(opMsg, "error", err)
		return huma.Error409Conflict(msgConflict)
	case errors.Is(err, domain.ErrUnauthorized):
		slog.Warn(opMsg, "error", err)
		return huma.Error401Unauthorized(msgUnauthorized)
	case errors.Is(err, domain.ErrForbidden):
		slog.Warn(opMsg, "error", err)
		return huma.Error403Forbidden(msgForbidden)
	case errors.Is(err, domain.ErrBadInput):
		slog.Warn(opMsg, "error", err)
		return huma.Error400BadRequest(msgBadRequest)
	case errors.Is(err, domain.ErrUnprocessableEntity):
		slog.Warn(opMsg, "error", err)
		return huma.Error422UnprocessableEntity(msgUnprocessableEntity)
	}

	slog.Error(opMsg, "error", err)
	return huma.Error500InternalServerError(opMsg)
}

// mapSlice returns a new slice with fn applied to every element of in.
// Always returns a non-nil slice
func mapSlice[T, U any](in []T, fn func(T) U) []U {
	out := make([]U, 0, len(in))
	for _, v := range in {
		out = append(out, fn(v))
	}

	return out
}
