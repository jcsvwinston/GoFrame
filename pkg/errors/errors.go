// Package errors provides domain-specific error types for the GoFrame framework.
// It defines a DomainError type with HTTP status codes and JSON serialization,
// along with convenience constructors for common error cases.
//
// The package supports Laravel-style error handling with:
// - Reportable exceptions (custom logging/reporting)
// - Renderable exceptions (custom HTTP responses)
// - Configurable log levels per error type
// - Global error context
// - Exception throttling
package errors

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
)

// Reportable is an interface that errors can implement to provide custom
// reporting logic (e.g., sending to Sentry, Flare, or other external services).
type Reportable interface {
	// Report handles the error reporting. Return false to fall back to default logging.
	Report(ctx context.Context, logger *slog.Logger) bool
}

// Renderable is an interface that errors can implement to provide custom
// HTTP rendering logic.
type Renderable interface {
	// Render returns a custom HTTP response. Return false to fall back to default rendering.
	Render(w http.ResponseWriter, r *http.Request) bool
}

// ContextProvider is an interface that errors can implement to provide
// additional context for logging.
type ContextProvider interface {
	// Context returns additional context data for logging.
	Context() map[string]any
}

// LogLevelProvider is an interface that errors can implement to specify
// their log level.
type LogLevelProvider interface {
	// LogLevel returns the log level for this error.
	LogLevel() slog.Level
}

// DomainError represents a structured application error with an HTTP status code,
// a machine-readable code, a human-readable message, and optional details.
type DomainError struct {
	Code       string `json:"code"`
	Message    string `json:"message"`
	StatusCode int    `json:"-"`
	Details    any    `json:"details,omitempty"`
}

// Error implements the error interface.
func (e *DomainError) Error() string {
	return e.Message
}

// WithDetails returns a copy of the error with the given details attached.
func (e *DomainError) WithDetails(details any) *DomainError {
	return &DomainError{
		Code:       e.Code,
		Message:    e.Message,
		StatusCode: e.StatusCode,
		Details:    details,
	}
}

// Report implements Reportable for DomainError.
// By default, DomainErrors are not reported (returns false to use default logging).
func (e *DomainError) Report(ctx context.Context, logger *slog.Logger) bool {
	return false // Use default logging
}

// Render implements Renderable for DomainError.
// By default, DomainErrors use the standard JSON rendering (returns false).
func (e *DomainError) Render(w http.ResponseWriter, r *http.Request) bool {
	return false // Use default rendering
}

// Context implements ContextProvider for DomainError.
// Returns the details as context for logging.
func (e *DomainError) Context() map[string]any {
	if e.Details == nil {
		return nil
	}
	if m, ok := e.Details.(map[string]any); ok {
		return m
	}
	return map[string]any{"details": e.Details}
}

// LogLevel implements LogLevelProvider for DomainError.
// Returns ERROR for 5xx, DEBUG for 4xx.
func (e *DomainError) LogLevel() slog.Level {
	if e.StatusCode >= 500 {
		return slog.LevelError
	}
	return slog.LevelDebug
}

// NotFound creates a 404 error indicating a resource was not found.
func NotFound(resource, id string) *DomainError {
	return &DomainError{
		Code:       "NOT_FOUND",
		Message:    fmt.Sprintf("%s '%s' not found", resource, id),
		StatusCode: http.StatusNotFound,
	}
}

// BadRequest creates a 400 error for malformed or invalid requests.
func BadRequest(message string) *DomainError {
	return &DomainError{
		Code:       "BAD_REQUEST",
		Message:    message,
		StatusCode: http.StatusBadRequest,
	}
}

// Unauthorized creates a 401 error for unauthenticated requests.
func Unauthorized(message string) *DomainError {
	return &DomainError{
		Code:       "UNAUTHORIZED",
		Message:    message,
		StatusCode: http.StatusUnauthorized,
	}
}

// Forbidden creates a 403 error for unauthorized access attempts.
func Forbidden(message string) *DomainError {
	return &DomainError{
		Code:       "FORBIDDEN",
		Message:    message,
		StatusCode: http.StatusForbidden,
	}
}

// Conflict creates a 409 error for resource conflicts.
func Conflict(message string) *DomainError {
	return &DomainError{
		Code:       "CONFLICT",
		Message:    message,
		StatusCode: http.StatusConflict,
	}
}

// InternalError creates a 500 error for unexpected server errors.
func InternalError(message string) *DomainError {
	return &DomainError{
		Code:       "INTERNAL_ERROR",
		Message:    message,
		StatusCode: http.StatusInternalServerError,
	}
}

// ValidationFailed creates a 422 error with per-field validation details.
func ValidationFailed(fields map[string]string) *DomainError {
	return &DomainError{
		Code:       "VALIDATION_FAILED",
		Message:    "validation failed",
		StatusCode: http.StatusUnprocessableEntity,
		Details:    fields,
	}
}
