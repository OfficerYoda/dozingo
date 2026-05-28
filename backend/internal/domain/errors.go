package domain

import "errors"

var (
	// ErrNotFound indicates a requested resource does not exist.
	ErrNotFound = errors.New("not found")

	// ErrConflict indicates a uniqueness or state conflict (e.g. duplicate key).
	ErrConflict = errors.New("conflict")

	// ErrUnauthorized indicates the caller is not authenticated.
	ErrUnauthorized = errors.New("unauthorized")

	// ErrForbidden indicates the caller is authenticated but lacks permission.
	ErrForbidden = errors.New("forbidden")

	// ErrBadInput indicates the input failed a domain rule beyond schema validation.
	ErrBadInput = errors.New("bad input")

	// ErrUnprocessableEntity indicates the input field could be parsed but fails validation.
	ErrUnprocessableEntity = errors.New("unprocessable entity")

	// ErrGone indicates the resource is no longer valid because it expired.
	ErrGone = errors.New("gone")
)
