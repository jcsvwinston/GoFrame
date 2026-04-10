# Error Handling Guide

Reference date: 2026-04-10.
Status: Current.

This guide covers GoFrame's error handling system (`pkg/errors`), including domain error types, HTTP status mapping, and custom error patterns.

## Table of Contents

- [Overview](#overview)
- [Domain Error Types](#domain-error-types)
- [Creating Errors](#creating-errors)
- [HTTP Error Responses](#http-error-responses)
- [Error-to-HTTP Status Mapping](#error-to-http-status-mapping)
- [Custom Error Domains](#custom-error-domains)
- [Error Wrapping and Unwrapping](#error-wrapping-and-unwrapping)
- [Admin Panel Error Handling](#admin-panel-error-handling)

---

## Overview

GoFrame provides a structured error system in `pkg/errors` that separates **domain errors** from **HTTP concerns**. This allows:

1. Consistent error codes across your application.
2. Automatic HTTP status code mapping.
3. Machine-readable error responses for API consumers.
4. Localized error messages (when combined with i18n).

---

## Domain Error Types

GoFrame defines standard domain errors:

| Error Code | HTTP Status | Description |
|------------|-------------|-------------|
| `ErrNotFound` | 404 | Resource not found |
| `ErrUnauthorized` | 401 | Authentication required |
| `ErrForbidden` | 403 | Insufficient permissions |
| `ErrValidation` | 422 | Request validation failed |
| `ErrConflict` | 409 | Resource conflict |
| `ErrInternal` | 500 | Internal server error |
| `ErrBadRequest` | 400 | Malformed request |
| `ErrRateLimited` | 429 | Rate limit exceeded |
| `ErrUnavailable` | 503 | Service temporarily unavailable |

---

## Creating Errors

### Using error constructors

```go
import "github.com/jcsvwinston/GoFrame/pkg/errors"

// Simple domain error
err := errors.NewNotFound("article", "42")
// Error message: "article not found: 42"

// Error with details
err := errors.NewValidation("email", "invalid format")
// Error message: "validation error: email: invalid format"

// Error with custom code
err := errors.New("CUSTOM_CODE", "Something went wrong", 400)
```

### Common constructors

```go
// Not Found
err := errors.NewNotFound(resource, identifier)

// Unauthorized
err := errors.NewUnauthorized("authentication required")

// Forbidden
err := errors.NewForbidden("admin access required")

// Validation
err := errors.NewValidation(field, message)
err := errors.NewValidationMulti(map[string]string{
    "email": "invalid format",
    "password": "too short",
})

// Conflict
err := errors.NewConflict("email already exists")

// Bad Request
err := errors.NewBadRequest("invalid JSON payload")

// Internal
err := errors.NewInternal("database connection failed")
err = errors.NewInternalErr(err) // Wraps underlying error

// Rate Limited
err := errors.NewRateLimited("try again in 60 seconds")

// Unavailable
err := errors.NewUnavailable("service is down for maintenance")
```

---

## HTTP Error Responses

### Writing errors to HTTP responses

```go
import "github.com/jcsvwinston/GoFrame/pkg/errors"

func GetArticle(w http.ResponseWriter, r *http.Request) {
    id := chi.URLParam(r, "id")

    article, err := articleService.Find(id)
    if err != nil {
        errors.WriteHTTP(w, err)
        return
    }

    ctx := router.NewContext(w, r)
    ctx.JSON(article)
}
```

### Error response format

```json
{
    "error": {
        "code": "NOT_FOUND",
        "message": "article not found: 42",
        "status": 404
    }
}
```

For validation errors:

```json
{
    "error": {
        "code": "VALIDATION_ERROR",
        "message": "request validation failed",
        "status": 422,
        "details": {
            "email": "invalid format",
            "password": "must be at least 8 characters"
        }
    }
}
```

---

## Error-to-HTTP Status Mapping

GoFrame automatically maps domain errors to HTTP status codes:

```go
// Internal mapping (pkg/errors/handler.go)
var statusCodeMap = map[string]int{
    "NOT_FOUND":        http.StatusNotFound,          // 404
    "UNAUTHORIZED":     http.StatusUnauthorized,      // 401
    "FORBIDDEN":        http.StatusForbidden,         // 403
    "VALIDATION_ERROR": http.StatusUnprocessableEntity, // 422
    "CONFLICT":         http.StatusConflict,          // 409
    "BAD_REQUEST":      http.StatusBadRequest,        // 400
    "INTERNAL_ERROR":   http.StatusInternalServerError, // 500
    "RATE_LIMITED":     http.StatusTooManyRequests,   // 429
    "UNAVAILABLE":      http.StatusServiceUnavailable, // 503
}
```

### Custom mapping

You can extend the mapping for your application:

```go
func init() {
    errors.RegisterStatusCode("PAYMENT_REQUIRED", http.StatusPaymentRequired)
}

err := errors.New("PAYMENT_REQUIRED", "subscription expired", 402)
errors.WriteHTTP(w, err) // Returns 402
```

---

## Custom Error Domains

Define application-specific error types:

```go
// internal/errors/errors.go
package errors

import gferrors "github.com/jcsvwinston/GoFrame/pkg/errors"

// Business domain errors
var (
    ErrArticleAlreadyPublished = gferrors.NewConflict("article is already published")
    ErrArticleDraftNotFound    = gferrors.NewNotFound("draft article", "")
    ErrInsufficientCredits     = gferrors.New("INSUFFICIENT_CREDITS", "not enough credits", 402)
)

// Validation errors
func NewArticleValidationError(field, message string) error {
    return gferrors.NewValidation(field, message)
}
```

### Error with metadata

```go
type DomainError struct {
    Code       string
    Message    string
    HTTPStatus int
    Meta       map[string]any
}

func (e *DomainError) Error() string {
    return e.Message
}

// Usage
err := &DomainError{
    Code:       "ARTICLE_PUBLISH_FAILED",
    Message:    "could not publish article",
    HTTPStatus: 422,
    Meta: map[string]any{
        "article_id": 42,
        "reason":     "validation_failed",
    },
}
```

---

## Error Wrapping and Unwrapping

GoFrame errors support Go 1.13+ error wrapping:

```go
import (
    "errors"
    "fmt"

    gferrors "github.com/jcsvwinston/GoFrame/pkg/errors"
)

// Wrap domain error with context
err := gferrors.NewNotFound("article", "42")
wrapped := fmt.Errorf("get article: %w", err)

// Unwrap
var notFound *gferrors.DomainError
if errors.As(wrapped, &notFound) {
    // Handle not found
}

// Check specific error type
if errors.Is(wrapped, gferrors.ErrNotFound) {
    // Handle not found
}
```

### Internal error wrapping

```go
// Wrap underlying infrastructure errors
dbErr := database.Query("SELECT ...")
if dbErr != nil {
    return gferrors.NewInternalErr(dbErr)
}

// The original error is preserved for debugging
// but the HTTP response shows the domain error
var internalErr *gferrors.InternalError
if errors.As(err, &internalErr) {
    log.Error("underlying error", "error", internalErr.Unwrap())
}
```

---

## Admin Panel Error Handling

The admin panel has its own error handling pattern:

```go
// Admin handlers convert auth errors to domain errors
func authErrorToDomain(err error) error {
    if err == nil {
        return nil
    }
    if errors.Is(err, auth.ErrSessionExpired) {
        return gferrors.NewUnauthorized("session expired")
    }
    if errors.Is(err, auth.ErrInvalidCredentials) {
        return gferrors.NewUnauthorized("invalid credentials")
    }
    return gferrors.NewInternalErr(err)
}
```

---

## Best Practices

1. **Use domain errors, not HTTP errors**: Define errors in your domain layer, map to HTTP in handlers.
2. **Wrap, don't swallow**: Always wrap underlying errors with context.
3. **Consistent codes**: Use the same error codes across your application.
4. **Log at the right level**: Log `INTERNAL` errors as `ERROR`, validation errors as `WARN`.
5. **Don't expose internals**: Never leak database errors or stack traces to API consumers.
6. **Use validation errors for user input**: Return field-level details for 422 responses.

```go
// Good
err := errors.NewNotFound("article", id)
errors.WriteHTTP(w, err)

// Bad (leaks internal details)
http.Error(w, dbError.Error(), 500)
```
