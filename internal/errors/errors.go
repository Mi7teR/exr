package errors

import "errors"

var (
	// ErrNotFound is returned when a resource is not found.
	ErrNotFound = errors.New("not found")
	// ErrInvalidArgument is returned when an argument is invalid.
	ErrInvalidArgument = errors.New("invalid argument")
	// ErrInternal is returned when an internal error occurs.
	ErrInternal = errors.New("internal error")
)
