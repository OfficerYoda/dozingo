package handler

import (
	"errors"
	"log/slog"

	"github.com/danielgtaylor/huma/v2"
	"github.com/officeryoda/dozingo/internal/domain"
)

func toHumaErr(err error, notFoundMsg, opMsg string) error {
	switch {
	case errors.Is(err, domain.ErrNotFound):
		if notFoundMsg == "" {
			notFoundMsg = err.Error()
		}
		return huma.Error404NotFound(notFoundMsg)
	case errors.Is(err, domain.ErrConflict):
		return huma.Error409Conflict(err.Error())
	case errors.Is(err, domain.ErrUnauthorized):
		return huma.Error401Unauthorized(err.Error())
	case errors.Is(err, domain.ErrForbidden):
		return huma.Error403Forbidden(err.Error())
	case errors.Is(err, domain.ErrInvalid):
		return huma.Error400BadRequest(err.Error())
	case errors.Is(err, domain.ErrUnprocessableEntity):
		return huma.Error422UnprocessableEntity(err.Error())
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
