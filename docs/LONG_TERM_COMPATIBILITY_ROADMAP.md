# Long-Term Compatibility Roadmap

Reference date: 2026-04-07.
Status: Current (strategic baseline before `v1.0`).

This roadmap defines how GoFrame will evolve so application code remains functional across framework upgrades, even when teams choose not to adopt new features immediately.

## North Star

From `v1.0` onward, upgrading GoFrame within `v1.x` must not force application code rewrites.

Allowed in `v1.x`:

- optional feature adoption
- new APIs added in backward-compatible form
- performance and security hardening without behavioral breakage

Not allowed in `v1.x`:

- removing or changing behavior of stable public APIs
- changing config semantics in a way that breaks existing projects
- forcing migrations in app code just to remain operational

## Compatibility Contract Scope

Stable from `v1.0`:

- public Go APIs in `pkg/*` documented as stable
- CLI command behavior and flag contracts marked stable
- `goframe.yaml` stable keys and defaults
- plugin envelope contract (`version: v1`) and capability semantics
- SQL model/tag conventions and migration lifecycle behavior

Out of stability scope (can evolve faster):

- `internal/*` packages
- explicitly experimental commands and flags
- undocumented/internal extension points

## Non-Negotiable Architecture Rules

1. Third-party dependency firewall
- Public APIs must not expose third-party concrete types.
- External libraries stay behind GoFrame-owned interfaces/adapters.
- Replacing router/DB/plugin backends must not require app code changes.

2. Compatibility-by-default over innovation-by-default
- New behaviors ship opt-in first (feature flags/capabilities).
- Default behavior cannot change in a breaking way inside `v1.x`.

3. Versioned contracts for integration boundaries
- Plugin SDK stays explicitly versioned (`v1`, future `v2`, etc.).
- Config schema is versioned and supports automatic upgrade paths.
- Data/migration contract changes require explicit compatibility notes.

4. Deprecation without forced migration
- Deprecate first, keep behavior until next major (`v2`).
- Provide shims/adapters and migration tooling before any removal.

## Delivery Model (Phased)

## Phase A: Foundations Freeze (`v0.6.x` -> `v0.7.x`)

Goals:

- freeze and document the stable public surface target for `v1.0`
- classify APIs/commands as `stable`, `transitional`, or `experimental`
- define explicit compatibility policy and release gates

Deliverables:

- Public API inventory document (owned packages + stability level)
- CLI command stability matrix
- Config key registry with stability tags
- dependency policy baseline (allowed critical dependencies + replacement criteria)

Exit criteria:

- every public entrypoint tagged with lifecycle state
- release checklist includes compatibility gate

## Phase B: Compatibility Harness (`v0.7.x` -> `v0.8.x`)

Goals:

- prove old app code continues to build/run on newer framework versions

Deliverables:

- fixture applications (at least 3 profiles: minimal API, admin-heavy, plugin-heavy)
- compatibility CI pipeline:
  - compile fixture apps against current framework
  - run smoke scenarios and compare golden outputs
- golden contract tests for:
  - CLI stable commands
  - config parsing/defaults
  - plugin request/response envelope

Exit criteria:

- no compatibility regression merges without explicit waiver
- repeatable compatibility report artifact per release candidate

## Phase C: Dependency Decoupling (`v0.8.x` -> `v0.9.x`)

Goals:

- ensure third-party upgrades do not leak into user code

Deliverables:

- adapter layers completed for critical components:
  - router
  - data access/driver integration
  - plugin runtime boundary
  - observability boundary
- remove public exposure of external concrete types
- add fail-fast tests ensuring no third-party type leaks in stable APIs

Exit criteria:

- dependency swap drill succeeds in at least one critical subsystem without app changes

## Phase D: `v1.0` Readiness Lock (`v0.9.x`)

Goals:

- freeze stable contracts before `v1.0`

Deliverables:

- contract freeze checklist signed:
  - public API surface freeze
  - plugin SDK `v1` freeze
  - config schema `v1` freeze
- migration assistant commands for deprecated transitional surfaces
- published compatibility SLO for `v1.x`

Exit criteria:

- two consecutive release candidates with green compatibility harness
- zero unresolved high-severity compatibility findings

## Phase E: `v1.x` Stewardship (post-`v1.0`)

Goals:

- preserve long-term trust and operational continuity

Operating rules:

- no removals of stable APIs in `v1.x`
- deprecations include:
  - replacement path
  - codemod or automated guide where possible
  - minimum deprecation window until `v2`
- security fixes and dependency upgrades are compatibility-validated

Compatibility SLO target:

- `>= 99%` fixture-application pass rate in release validation
- `0` known critical compatibility regressions at GA release

## Plugin and Extension Strategy (Long Horizon)

- keep Plugin SDK `v1` stable for all `v1.x`
- add capability negotiation instead of behavior rewrites:
  - framework asks plugin what it supports
  - plugin can opt into newer optional fields without breaking old behavior
- preserve legacy bridges while adoption migrates

## Data Layer Strategy (Long Horizon)

- application models stay framework-owned and database-agnostic by default
- engine-specific optimizations stay opt-in via capabilities/options
- migration behavior stays deterministic and backward-compatible
- multi-engine support is promoted from exploratory to required only after:
  - live CI stability
  - critical command parity
  - compatibility-harness evidence

## Governance and Release Gates

Every release candidate must include:

- compatibility harness report
- dependency impact report (critical dependencies only)
- explicit statement:
  - `no breaking changes`, or
  - `breaking changes (major-only)` with migration plan

Required docs to keep in sync:

- `docs/VERSIONING.md`
- `docs/GO_VERSION_POLICY.md`
- `docs/PLUGIN_SDK.md`
- `docs/CI_MATRIX.md`
- `CHANGELOG.md`

## Initial Actions (Next 4 Weeks)

1. Create public API inventory + stability tags.
2. Build first fixture-app compatibility harness in CI.
3. Add compatibility gate to release checklist and CI required checks.
4. Publish first dependency impact report template.
5. Define deprecation template (announcement, timeline, migration path, tooling).

This roadmap is the strategic contract for the path to `v1.0` and ongoing `v1.x` maintenance.
