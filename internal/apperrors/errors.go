package apperrors

import "errors"

var (
	ErrNotFound     = errors.New("not found")
	ErrConflict     = errors.New("conflict")
	ErrForbidden    = errors.New("forbidden")
	ErrInvalidState = errors.New("invalid state")
	ErrUnauthorized = errors.New("unauthorized")
	ErrIdempotency  = errors.New("idempotency conflict")
)
