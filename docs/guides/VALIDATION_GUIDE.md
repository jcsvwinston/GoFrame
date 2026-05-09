# Validation Guide

Reference date: 2026-04-10.
Status: Current.

This guide covers Nucleus's validation system (`pkg/validate`), including built-in rules, custom validators, and error handling patterns.

## Table of Contents

- [Overview](#overview)
- [Struct Tag Validation](#struct-tag-validation)
- [Built-in Rules](#built-in-rules)
- [Custom Rules](#custom-rules)
- [Validation Error Handling](#validation-error-handling)
- [Request Validation in Handlers](#request-validation-in-handlers)
- [Nested Validation](#nested-validation)
- [Conditional Validation](#conditional-validation)

---

## Overview

Nucleus uses `github.com/go-playground/validator/v10` as its validation engine, exposed through `pkg/validate`. Validation is triggered via struct tags and supports:

- Field-level validation rules
- Cross-field validation
- Custom validation functions
- Localized error messages

---

## Struct Tag Validation

Define validation rules using struct tags:

```go
import "github.com/jcsvwinston/nucleus/pkg/validate"

type CreateArticleRequest struct {
    Title    string   `validate:"required,min=3,max=200"`
    Slug     string   `validate:"required,alphanumdash"`
    Body     string   `validate:"required,min=10"`
    Status   string   `validate:"required,oneof=draft published archived"`
    Category string   `validate:"omitempty,uuid"`
    Tags     []string `validate:"dive,min=2,max=50"`
}

func (h *ArticleHandler) Create(w http.ResponseWriter, r *http.Request) {
    var req CreateArticleRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        errors.WriteHTTP(w, errors.NewBadRequest("invalid JSON"))
        return
    }

    if err := validate.Struct(&req); err != nil {
        errors.WriteHTTP(w, errors.NewValidationMulti(err.FieldErrors()))
        return
    }

    // Process valid request
}
```

---

## Built-in Rules

Nucleus exposes all validator/v10 rules. The most commonly used:

### String Rules

| Rule | Description | Example |
|------|-------------|---------|
| `required` | Field must be non-zero | `validate:"required"` |
| `min=N` | Minimum length | `validate:"min=3"` |
| `max=N` | Maximum length | `validate:"max=200"` |
| `len=N` | Exact length | `validate:"len=32"` |
| `email` | Valid email format | `validate:"required,email"` |
| `url` | Valid URL format | `validate:"url"` |
| `uuid` | Valid UUID format | `validate:"uuid"` |
| `oneof=A B C` | Must be one of the values | `validate:"oneof=draft published"` |

### Number Rules

| Rule | Description | Example |
|------|-------------|---------|
| `gt=N` | Greater than N | `validate:"gt=0"` |
| `gte=N` | Greater than or equal to N | `validate:"gte=0"` |
| `lt=N` | Less than N | `validate:"lt=100"` |
| `eq=N` | Equal to N | `validate:"eq=18"` |

### Slice/Map Rules

| Rule | Description | Example |
|------|-------------|---------|
| `dive` | Apply rules to each element | `validate:"dive,min=1"` |
| `min=N` | Minimum length | `validate:"min=1"` |
| `unique` | No duplicate elements | `validate:"unique"` |

---

## Custom Rules

### Registering custom validators

```go
import (
    "regexp"

    "github.com/go-playground/validator/v10"
    "github.com/jcsvwinston/nucleus/pkg/validate"
)

func init() {
    alphanumDashRegex := regexp.MustCompile(`^[a-zA-Z0-9\-]+$`)
    validate.RegisterValidation("alphanumdash", func(fl validator.FieldLevel) bool {
        return alphanumDashRegex.MatchString(fl.Field().String())
    })
}
```

---

## Validation Error Handling

```go
func handleValidationError(err error) error {
    var validationErrs validator.ValidationErrors
    if !errors.As(err, &validationErrs) {
        return errors.NewInternalErr(err)
    }

    fieldErrors := make(map[string]string)
    for _, fe := range validationErrs {
        fieldErrors[fe.Field()] = formatValidationError(fe)
    }

    return errors.NewValidationMulti(fieldErrors)
}
```

---

## Request Validation in Handlers

```go
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
    var req CreateRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        errors.WriteHTTP(w, errors.NewBadRequest("invalid JSON"))
        return
    }

    if err := validate.Struct(&req); err != nil {
        errors.WriteHTTP(w, handleValidationError(err))
        return
    }

    // Process valid request
}
```

---

## Nested Validation

```go
type Address struct {
    Street  string `validate:"required,min=5"`
    City    string `validate:"required,min=2"`
    Country string `validate:"required,len=2"`
}

type UserProfile struct {
    Name    string  `validate:"required"`
    Email   string  `validate:"required,email"`
    Address Address `validate:"required,dive"`
}
```

---

## Conditional Validation

### Using `omitempty`

```go
type UpdateRequest struct {
    Title  string `validate:"omitempty,min=3"`
    Status string `validate:"omitempty,oneof=draft published"`
}
```

### Using `required_without`

```go
type ContactRequest struct {
    Email string `validate:"required_without=Phone"`
    Phone string `validate:"required_without=Email"`
}
```
