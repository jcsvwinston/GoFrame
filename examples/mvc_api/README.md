# MVC + API Example

This example demonstrates how to run a GoFrame project that combines:

- MVC pages
- REST API endpoints
- embedded admin panel

## Run

From repository root:

```bash
go run ./examples/mvc_api
```

Open:

- `http://localhost:8080/`
- `http://localhost:8080/api/health`
- `http://localhost:8080/admin`

## Purpose

Use this example as a reference for:

- app bootstrap with `pkg/app`
- model registration and admin exposure
- route composition and practical wiring
