// Package quark provides a modern, type-safe ORM for Go.
// It supports multiple SQL dialects and is designed to be framework-agnostic.
package quark

import "errors"

// Common errors returned by quark operations.
var (
	// ErrNotFound indicates that no record was found for the given criteria.
	ErrNotFound = errors.New("record not found")

	// ErrInvalidModel indicates that the provided model is invalid or not registered.
	ErrInvalidModel = errors.New("invalid model")

	// ErrInvalidQuery indicates that the query is malformed or invalid.
	ErrInvalidQuery = errors.New("invalid query")

	// ErrInvalidIdentifier indicates that a table or column identifier is invalid.
	ErrInvalidIdentifier = errors.New("invalid identifier")

	// ErrDialectNotSupported indicates that the database dialect is not supported.
	ErrDialectNotSupported = errors.New("dialect not supported")

	// ErrConnection indicates a database connection error.
	ErrConnection = errors.New("database connection error")

	// ErrTimeout indicates that a query timed out.
	ErrTimeout = errors.New("query timeout")

	// ErrConstraintViolation indicates a database constraint violation.
	ErrConstraintViolation = errors.New("constraint violation")
)
