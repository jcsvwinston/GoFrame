---
sidebar_position: 2
title: Quickstart
---

# Quickstart

Five minutes from zero to a running app with a database, a model, and an
embedded admin panel.

## 1 — Scaffold a project

```bash
nucleus new myapp
cd myapp
go mod tidy
```

`nucleus new` writes a self-contained Go module. There is no `replace`
directive, no local clone of Nucleus required.

## 2 — Run the server

```bash
nucleus serve
```

Four endpoints are now live:

| URL                                | Purpose                          |
| ---------------------------------- | -------------------------------- |
| `http://localhost:8080/`           | The web app                      |
| `http://localhost:8080/api/...`    | Auto-mounted REST endpoints      |
| `http://localhost:8080/admin`      | Embedded admin panel             |
| `http://localhost:8080/healthz`    | Liveness/readiness checks        |

The default config (`nucleus.yml`) uses SQLite at `app.db`. Override the
database with environment variables or by editing `nucleus.yml`.

## 3 — A minimal API in code

:::note Phase 1 API — full example coming in v0.9.X

The `pkg/nucleus` entry point was rewritten in ADR-010 Phase 1 (landed
2026-05-16). The legacy fluent chain (`Port`, `SQLite`, `AutoMigrate`,
`Get`, `Run`, …) is removed. The new Phase 1 surface uses
`nucleus.New().Use(...).Mount(...).Start()` with `Module`-based
registration.

The canonical API is documented in the
[`pkg/nucleus` godoc](https://github.com/jcsvwinston/nucleus/blob/main/pkg/nucleus/nucleus.go)
and in [ADR-010](https://github.com/jcsvwinston/nucleus/blob/main/docs/adrs/ADR-010-fluent-api-v2-pkg-nucleus.md). A
worked single-file example will land in Phase 4 / v0.9.X.

For a self-contained runnable app today, use the scaffolded project from
steps 1–2 above or the `pkg/app`-based pattern in
[Concepts → Application](../concepts/application.md).

:::

:::info AutoMigrate (dev-mode only)

`(*app.App).AutoMigrate(models ...any)` derives idempotent
`CREATE TABLE` statements from struct tags and runs them against the
configured database. Five dialects are supported: **SQLite, PostgreSQL,
MySQL, MSSQL, and Oracle** — each via its own deterministic scaffold
builder in
[`pkg/model`](https://github.com/jcsvwinston/nucleus/blob/main/pkg/model).
On SQLite/Postgres/MySQL the generated SQL uses `CREATE TABLE IF NOT
EXISTS`; on MSSQL it wraps the CREATE in `IF OBJECT_ID(..., 'U') IS
NULL`; on Oracle it wraps it in a PL/SQL block that swallows `ORA-00955`
("name is already used by an existing object"). Either way the operation
is safe to re-run.

`AutoMigrate` returns `db.ErrAutoMigrate` only for unknown drivers.

`AutoMigrate` does **not** alter existing tables — it is
`CREATE IF NOT EXISTS` only. For production schema evolution, prefer
explicit SQL migration files (`migrations/*.up.sql` plus
`nucleus migrate`): they are reversible, reviewable in PR diffs, and the
only path the framework offers compatibility guarantees on.
`nucleus migrate drift` will surface any applied migration that has since
lost its `.up.sql` file on disk.

:::

## 4 — Run a migration

For non-trivial apps, write SQL migrations under `migrations/` and apply
them with the CLI:

```bash
nucleus migrate         # apply pending migrations
nucleus migrate status  # show plan vs. applied
nucleus migrate down    # roll back the most recent batch
```

## 5 — Create an admin user

```bash
nucleus createuser
```

Prompts for username, email and password. The user goes into the auth
table referenced by your `nucleus.yml`. You can now sign in to the admin
panel at `/admin`.

## Next steps

- **[Project structure](./project-structure.md)** — how a scaffolded
  project is laid out.
- **[Concepts → Application](../concepts/application.md)** — how the
  application container is wired up (`pkg/app` and `pkg/nucleus`).
- **[Concepts → Configuration](../concepts/configuration.md)** — the
  `nucleus.yml` schema.
