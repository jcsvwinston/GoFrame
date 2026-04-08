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

- `http://localhost:8090/`
- `http://localhost:8090/api/health`
- `http://localhost:8090/admin`

## Purpose

Use this example as a reference for:

- app bootstrap with `pkg/app`
- model registration and admin exposure
- route composition and practical wiring

## Live Feature Flag Demo

This example registers a runtime flag:

- `articles_preview_mode` (default `false`)

And exposes a demo endpoint:

- `GET /api/articles/live-flag`

Behavior:

- `false` -> mode `published_only` (drafts are hidden)
- `true` -> mode `preview_all` (drafts are included)

How to test manually:

1. Create one draft article (`published=false`) from `/admin` or `POST /api/articles`.
2. Call `GET /api/articles/live-flag` and verify draft is hidden.
3. Go to `/admin/system` -> **Live feature flags** -> enable `articles_preview_mode`.
4. Call `GET /api/articles/live-flag` again and verify draft is now visible.
