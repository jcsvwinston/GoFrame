# Architecture Decision Records

Reference date: 2026-04-10.
Status: Current.

This directory contains Architecture Decision Records (ADRs) documenting key technical choices in GoFrame.

---

## ADR-001: stdlib-First Runtime Design

**Status:** Accepted
**Date:** 2026-03-01
**Superseded:** No

### Context

Go's standard library provides robust, production-ready building blocks: `net/http` for HTTP, `database/sql` for database access, `log/slog` for structured logging, and `context` for request lifecycle. However, many Go web frameworks wrap or replace these with third-party alternatives, creating:

1. **Dependency lock-in**: Applications become tied to framework-specific abstractions.
2. **Upgrade friction**: Framework updates break when underlying stdlib changes.
3. **Learning curve**: Developers must learn framework-specific patterns instead of portable Go idioms.
4. **Debugging complexity**: Stack traces include framework layers that obscure root causes.

### Decision

GoFrame is designed **stdlib-first**:

- **HTTP**: Uses `net/http` directly with a custom lightweight router (`pkg/router`). No Chi, Gin, or Echo as runtime dependencies.
- **Database**: Uses `database/sql` directly with driver-specific connection strings. No ORM abstraction (no GORM, Bun, or universal SQL layer).
- **Logging**: Uses `log/slog` (Go 1.21+) with OpenTelemetry integration. No zap, zerolog, or logrus.
- **Context**: Uses `context.Context` throughout for request lifecycle, cancellation, and value passing.

Third-party libraries are used **only when stdlib doesn't provide equivalent functionality**:

| Need | stdlib Gap | Chosen Library |
|------|-----------|----------------|
| Configuration | No YAML/env parsing | `koanf/v2` |
| JWT | No JWT support | `jwt/v5` |
| Sessions | No session management | `scs/v2` |
| Authorization | No policy engine | `casbin/v2` |
| Input validation | No struct tag validation | `validator/v10` |
| Background jobs | No job queue | `hibiken/asynq` (Redis-backed) |
| OpenTelemetry | No OTLP exporter | `otel/*` SDK packages |

### Consequences

#### Positive

- **Portability**: GoFrame applications use idiomatic Go patterns transferable to any Go project.
- **Debuggability**: Stack traces are shallow; errors originate from stdlib or application code.
- **Upgrade safety**: Go version bumps don't break framework internals.
- **Smaller dependency tree**: Fewer transitive dependencies mean fewer security vulnerabilities and faster builds.
- **Predictable behavior**: stdlib behavior is stable and well-documented.

#### Negative

- **More framework code**: GoFrame must implement routing, middleware chains, and request context helpers that other frameworks get from dependencies.
- **Reinvention risk**: Custom router/middleware must be thoroughly tested to match battle-tested alternatives.
- **Feature gap**: Some advanced features (e.g., automatic OpenAPI generation) may require third-party integration work.

### Compliance

All new code in GoFrame must:

1. Prefer stdlib packages over third-party alternatives when functionally equivalent.
2. Justify any new runtime dependency in the PR description.
3. Document the stdlib gap that the dependency fills.

---

## ADR-002: Django-Inspired CLI Design

**Status:** Accepted
**Date:** 2026-04-01
**Superseded:** No

### Context

GoFrame targets developers who value:

1. **Convention over configuration**: Predictable project structure and command patterns.
2. **Operational completeness**: All lifecycle tasks available from a single CLI.
3. **Migration safety**: SQL-first migrations with explicit control over schema changes.
4. **Admin productivity**: Embedded admin UI for rapid data management.

Django's `manage.py` has proven these patterns effective for long-lived systems over 20+ years.

### Decision

GoFrame's CLI (`goframe`) adopts Django-inspired command naming and workflow:

| Django Command | GoFrame Equivalent | Purpose |
|---------------|-------------------|---------|
| `runserver` | `serve` (alias: `runserver`) | Start development server |
| `startproject` | `new` (alias: `startproject`) | Create new project |
| `startapp` | `startapp` | Create new app module |
| `makemigrations` | `migrate create` (alias: `makemigrations`) | Create migration files |
| `showmigrations` | `migrate status` (alias: `showmigrations`) | Show migration status |
| `migrate` | `migrate` | Apply migrations |
| `createsuperuser` | `createuser` (alias: `createsuperuser`) | Create admin user |
| `dbshell` | `shell` (alias: `dbshell`) | SQL shell |
| `dumpdata` | `dumpdata` | Export data as fixtures |
| `loaddata` | `loaddata` | Import fixtures |
| `diffsettings` | `diffsettings` | Show config differences |
| `collectstatic` | `collectstatic` | Collect static files |
| `makemessages` | `makemessages` | Extract translatable strings |
| `compilemessages` | `compilemessages` | Compile translations |

### Consequences

#### Positive

- **Familiar workflow**: Django developers feel productive immediately.
- **Operational parity**: All lifecycle tasks available through consistent interface.
- **Alias flexibility**: Both Go-native and Django-style names supported.

#### Negative

- **Expectation mismatch**: Some Django commands have Go-specific implementations that differ subtly.
- **Command surface**: Large CLI requires maintenance and testing overhead.

### Compliance

New CLI commands must:

1. Follow the flat command-spec dispatch pattern in `internal/cli/root.go`.
2. Support global output flags (`--output`, `--color`, `--symbols`, `--json`).
3. Include production guardrails for destructive operations.
4. Be documented in `docs/CLI_CONTRACT_MATRIX.md` with lifecycle tags.
