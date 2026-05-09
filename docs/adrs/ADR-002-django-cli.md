# ADR-002: Django-Inspired CLI Design

**Status:** Accepted
**Date:** 2026-04-01
**Superseded:** No

## Context

GoFrame targets developers who value:

1. **Convention over configuration**: Predictable project structure and command patterns.
2. **Operational completeness**: All lifecycle tasks available from a single CLI.
3. **Migration safety**: SQL-first migrations with explicit control over schema changes.
4. **Admin productivity**: Embedded admin UI for rapid data management.

Django's `manage.py` has proven these patterns effective for long-lived systems over 20+ years.

## Decision

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

## Consequences

### Positive

- **Familiar workflow**: Django developers feel productive immediately.
- **Operational parity**: All lifecycle tasks available through consistent interface.
- **Alias flexibility**: Both Go-native and Django-style names supported.

### Negative

- **Expectation mismatch**: Some Django commands have Go-specific implementations that differ subtly.
- **Command surface**: Large CLI requires maintenance and testing overhead.

## Compliance

New CLI commands must:

1. Follow the flat command-spec dispatch pattern in `internal/cli/root.go`.
2. Support global output flags (`--output`, `--color`, `--symbols`, `--json`).
3. Include production guardrails for destructive operations.
4. Be documented in `docs/reference/CLI_CONTRACT_MATRIX.md` with lifecycle tags.
