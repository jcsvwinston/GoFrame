# GoFrame Technical Specification

Reference date: 2026-04-07.
Status: Current baseline (v0.6.x line).

This document defines the current, implemented technical baseline for GoFrame.
It replaces older design notes that referenced superseded architecture choices.

## 1. Scope and Precedence

This specification is implementation-first.

When documents conflict, precedence is:

1. `README.md`
2. Contract/governance docs in `docs/`:
- `docs/API_CONTRACT_INVENTORY.md`
- `docs/CLI_CONTRACT_MATRIX.md`
- `docs/CONFIG_KEY_REGISTRY.md`
- `docs/COMPATIBILITY_SLO.md`
3. This file (`SPEC.md`)
4. Detailed tutorials/manual examples

## 2. Core Principles

1. stdlib-first runtime design (`net/http`, `database/sql`, `log/slog`, `context`).
2. Explicit configuration and lifecycle; no hidden global singletons.
3. Compatibility by contract for stable API/CLI/config surfaces.
4. Security-by-default posture for production-sensitive features.
5. SQL-first operations and tooling, with deterministic CLI behavior.

## 3. Runtime Architecture

## 3.1 Application Container (`pkg/app`)

`app.New` wires and validates:

- config loading/normalization (`pkg/app/config.go`)
- logger (`pkg/observe`)
- SQL database map by alias (`database_default` + `databases.<alias>`)
- mail sender (`pkg/mail`)
- session manager (`pkg/auth`) with selected store (`memory|sql|redis`)
- HTTP router and middleware (`pkg/router`)
- request scope resolver for MultiSite/MultiTenant (`pkg/app/requestscope.go`)
- model registry (`pkg/model`)
- embedded admin panel (`pkg/admin`)

`App` exposes:

- `DB` (primary alias) and `DBs` (all opened aliases)
- `Database(alias)` and `DatabaseForRequest(r)` helpers
- graceful `Run`/`Shutdown` with shutdown hooks

## 3.2 HTTP and Middleware (`pkg/router`)

GoFrame uses its own router/mux abstractions (not Chi as a runtime dependency):

- route registration + mounting
- request middleware chain
- JSON helpers and HTTP utilities
- CORS/CSRF middleware
- rate limiting (`rate_limit_*`)
- OpenTelemetry HTTP instrumentation

## 3.3 Data and Model Layer

`pkg/db`:

- `database/sql`-based DB wrapper
- health checks and telemetry
- SQL migration executor and helpers

`pkg/model`:

- model metadata extraction from tags
- registry for app/admin integration
- generic CRUD operator
- metadata-driven migration scaffold generation
- model contract features include PK/FK/index metadata (simple + composite)

## 3.4 Admin (`pkg/admin`)

Embedded admin panel provides:

- model listing + schema endpoint
- CRUD API
- list/search/filter/order pagination
- CSV export and bulk actions
- action-level authorization hooks (`AdminAuth.Authorize`)
- session inventory endpoint and UI telemetry

## 3.5 Auth/Authz (`pkg/auth`, `pkg/authz`)

- JWT helpers
- password hashing helpers
- session manager with store backends:
- memory
- SQL table store
- Redis store
- session runtime metadata enrichment (`pod/host/instance`)
- Casbin integration points for authorization enforcement

## 3.6 Mail and Plugins (`pkg/mail`, `pkg/plugins`)

Mail:

- drivers: `noop`, `smtp`, `sendgrid`
- capability-style external provider bridge

Plugin runtime:

- provider discovery and capability schema handling
- `goframe-plugin-<provider>` primary external naming
- `goframe-mail-<driver>` legacy compatibility fallback

## 3.7 Background Tasks and Observability

Tasks (`pkg/tasks`):

- Asynq manager and worker runtime
- enqueue/process instrumentation hooks

Observability (`pkg/observe`):

- `slog` logger setup
- OpenTelemetry setup and shutdown

## 4. Dependency Reality (from `go.mod`)

Direct runtime dependencies include:

- Configuration: `koanf` (`v2` + yaml/env/file/struct providers)
- Auth/session/security: `jwt/v5`, `scs/v2`, `casbin/v2`, `validator/v10`, `x/crypto`
- SQL drivers: `modernc.org/sqlite`, `pgx/v5`, `go-sql-driver/mysql`
- Enterprise exploratory SQL drivers: `go-mssqldb`, `go-ora/v2`
- Redis: `go-redis/v9`
- Tasks: `hibiken/asynq`
- Observability: OpenTelemetry SDK/exporters

Not present as current runtime dependencies:

- Chi router
- Bun ORM/migrate
- GORM
- MongoDB driver

## 5. Configuration Contract (Current)

Canonical DB configuration is alias-based only:

```yaml
database_default: default
databases:
  default:
    url: sqlite://goframe.db
  analytics:
    url: postgres://...
```

Legacy single-URL DB keys are removed from the active contract.

Key contract families:

- server/runtime: `host`, `port`, timeouts, `env`, `debug`
- databases: `database_default`, `databases.<alias>.*`
- multisite: `multisite.*`
- multitenant: `multitenant.*`
- auth/session: `jwt_*`, `session_*`
- admin: `admin_prefix`, `admin_title`
- mail: `mail_driver`, `smtp_*`, `sendgrid_*`, `mail_from`
- security/rate limit: `rate_limit_*`
- i18n/static/storage: `default_locale`, `locales_path`, `static_*`, `storage_*`
- observability: `log_*`, `otlp_endpoint`, `metrics_path`

Reference registry: `docs/CONFIG_KEY_REGISTRY.md`.

## 6. MultiSite/MultiTenant Contract

MultiSite and MultiTenant are request-scope features in `pkg/app`.

- site resolution supports exact and wildcard host mapping
- tenant resolution supports `subdomain` and `header`
- tenant-to-database alias routing supports explicit mapping and templates
- security default: `multitenant.require_isolated_db: true`

Isolation guardrail behavior:

- startup validation rejects multi-tenant mappings that would share the same DB alias
- request routing rejects shared site DB alias fallback when tenant isolation is required

## 7. CLI Contract Baseline (`cmd/goframe`, `internal/cli`)

GoFrame ships stable operational CLI coverage for:

- runtime and diagnostics (`serve`, `routes`, `health`)
- scaffolding (`new`, `startapp`, `generate`)
- migrations and SQL maintenance
- data import/export/introspection
- auth/admin maintenance commands
- plugin and mail diagnostics
- static/i18n workflows
- test workflows and fixture server

Global output contract:

- `--output plain|pretty|json`
- `--color auto|always|never`
- `--symbols|--no-symbols`
- `--json` shorthand

Critical maintenance commands follow homogeneous output modes including structured JSON status payloads.

Reference lifecycle matrix: `docs/CLI_CONTRACT_MATRIX.md`.

## 8. Compatibility Governance

Stable compatibility is governed by:

- API inventory lifecycle tags (`docs/API_CONTRACT_INVENTORY.md`)
- CLI lifecycle matrix (`docs/CLI_CONTRACT_MATRIX.md`)
- config key registry lifecycle tags (`docs/CONFIG_KEY_REGISTRY.md`)
- compatibility SLO (`docs/COMPATIBILITY_SLO.md`)

Automated controls:

- stable contract freeze tests (`contracts/` + `scripts/ci/check_contract_freeze.sh`)
- compatibility harness (`scripts/ci/run_compatibility_harness.sh`)
- release compatibility report generation (`scripts/release/generate_compatibility_report.sh`)

## 9. Release-Readiness Baseline

Minimum release checks:

```bash
go test ./...
bash scripts/ci/check_contract_freeze.sh
bash scripts/ci/run_compatibility_harness.sh --enforce-threshold
bash scripts/release/generate_compatibility_report.sh --output dist/reports/compatibility_report.md --enforce-threshold
bash scripts/release/generate_dependency_impact_report.sh --output dist/reports/dependency_impact_report.md
```

Full rehearsal path:

```bash
bash scripts/release/rehearse_rc.sh
```

Checklist reference: `docs/RELEASE_CHECKLIST.md`.

## 10. Current Explicit Non-Goals

1. No universal ORM abstraction spanning SQL/document/cache.
2. No hidden auto-migrations at runtime.
3. No promise that all exploratory SQL engines are first-class stable contracts.
4. No silent breaking changes on stable surfaces inside a minor/patch line.
