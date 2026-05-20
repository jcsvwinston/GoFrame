# Current Iteration

> Owned by `session-curator`. Edited by other subagents only via the
> Session Start / Session End protocols (`CLAUDE.md` §2 and §5).

## Goal

No active iteration. Last completed: **freeze-scanner constructor-gap
fix** (PR #77 → `28f75b2`), archived at
`docs/iterations/2026-05-20-contract-freeze-scanner-constructors.md`.
The API freeze scanner now tracks `NewXxx` constructor functions
(previously invisible because `go/doc` files them under the returned
type's `Funcs`, not the package-level `docPkg.Funcs`); baseline
reseeded +78/−0.

ADR-010 §2 ("Config loading + merge engine") is feature-complete
(Phases 2a #73 / 2b #74 / 2c #75 / 2d #76). Next ADR-010 phase is
**Phase 3** (`/_/config` + `nucleus config print --effective`).

## Scope

- in: (TBD — owner to confirm scope when iteration is registered)
- out: (TBD)

## Acceptance criteria

- [ ] (TBD)

## Status

### Done (2026-05-20)

- **Freeze-scanner constructor-gap fix** (PR #77 → `28f75b2`).
  `contracts/freeze_test.go` now iterates `go/doc`'s per-type
  `typ.Funcs`; +78 constructor entries seeded into the API freeze
  baseline. Closes the Phase 2d gap (`NewMigrator` /
  `NewModuleMigrator` now frozen) and the broader class. Governance
  tooling only — no CHANGELOG, no semver bump.
- **ADR-010 §2 config loader feature-complete** (Phases 2b/2c/2d,
  PRs #74/#75/#76) — archived at
  `docs/iterations/2026-05-20-adr010-phase2bcd-config-loader-completion.md`.

### Done (earlier — see prior archives)

- v0.7.0 (PRs #56–#59); CSRF hardening (ADR-006); slog redaction
  (ADR-007); CSRF follow-ups + schema drift (ADR-008 + ADR-009);
  MSSQL/Oracle SchemaDrift (#66); pkg/app+pkg/nucleus inventory (#65);
  ADR-010 Phase 1 + examples purge (#71); ADR-010 Phase 2a (#73).

### In progress

- (none)

### Blocked

- (none)

## Candidate next steps (priority order, pending owner confirmation)

1. **`pkg/admin` bootstrap users-table DDL — dialect-aware fix for
   MSSQL/Oracle.** Real bug discovered during PR #66 CI: MSSQL
   `Incorrect syntax near 'nucleus_admin_users'`, Oracle
   `ORA-03076: unexpected item DEFAULT`. Still blocks
   `TestSQLMatrix_AutoMigrate_Exploratory` from re-enablement. Fix
   replicates the dialect-aware discipline of
   `pkg/model/migration_scaffold_{mssql,oracle}.go`. After the fix,
   re-wire `TestSQLMatrix_AutoMigrate_Exploratory` into
   `.github/workflows/ci.yml`. Could fold in candidate #4
   (`session_cookie_secure`).

2. **Scanner package-coverage gap** (opened by the 2026-05-20
   constructor-gap fix). `contracts/freeze_test.go`'s `packages` slice
   covers 15 `pkg/*` packages but OMITS six: `pkg/admin`,
   `pkg/circuit`, `pkg/health`, `pkg/observability`, `pkg/openapi`,
   `pkg/outbox` — their exported surface has zero removal-protection.
   Audit each and decide which to add, confirming lifecycle posture
   first (the flat baseline does not encode lifecycle):
   `pkg/outbox.NewKafkaBridge` is deliberately unfinished and must NOT
   be frozen until Kafka delivery lands; `pkg/openapi` is
   `experimental`; `pkg/admin` and `pkg/outbox` are `transitional`. A
   code comment at the `packages` slice now documents the deliberate
   omission.

3. **ADR-010 Phase 3 — `/_/config` + `nucleus config print
   --effective`.** Compliance items #6, #12, #13. Auth-gated by
   `WithAdmin()` (Casbin default-deny per ADR-004); redaction via
   `observe.DefaultRedactedKeys()` (ADR-007). Requires per-key source
   tracking the Phase 2 loader does not yet capture — the substantive
   new work.

4. **`session_cookie_secure` default `false`** (Phase 2b security-
   auditor MED-1). Pre-existing security default; the non-nullable
   mechanism doesn't cover it because the default is already
   permissive. Flip to `true` (breaking for local-dev plain HTTP) or
   add to the non-nullable set. Could fold into candidate #1.

5. **ADR-010 §2 layer 3 — field-semantic validation.** Ranges, enums,
   parseable durations (ADR-010 §96 validation layer 3). Out of the
   four-phase slicing; standalone follow-up on the now-complete merge
   engine.

6. **ADR-010 Phase 4 — Docs-sync + website + new reference
   applications under a freshly-scoped `examples/`.** Target: v0.9.X.
   New examples authored, website docs rewritten, manifest pattern
   introduced, compatibility-harness fixture profiles (`minimal-api`,
   `admin-heavy`, `plugin-heavy`) restored.

7. **Cloud Secrets Provider plugin extraction (AWS → GCP → Azure →
   Vault).** Three-iteration project following the SendGrid precedent
   (DEP-2026-002 / MA-2026-002). Removes AWS SDK from core `go.mod`.

8. **Column-type comparison in `SchemaDrift`.** Cross-dialect
   type-family compatibility table. Additive to `ExpectedColumn`.

9. **SchemaDrift end-to-end usage guide.** Bridge `model.ExtractMeta`
   → `[]db.ExpectedTable` documented in
   `docs/guides/MODELING_MULTI_DATABASE.md`.

10. **`go mod tidy` unblock.** Fix the `admin/proto` replace-directive
    issue so AWS SDK modules carry correct annotations (or, more
    elegantly, are gone entirely once candidate #7 lands). NOTE: the
    Phase 2b/2c koanf sub-modules (toml/v2, json, confmap) are stuck
    as `// indirect` until this unblocks.

11. **`tasks.Manager` struct→interface DEP** — optional DEP-2026-004
    for the binary-incompatible type-identity change.

12. **Audit §7 menores** — 503 path test for `/healthz`,
    endpoints-parity doc-parsing, individual tests for
    `pkg/health/{db,redis,storage}.go`.

## Carry-forward follow-ups (ADR-010 Phase 1, still open)

Non-blocker findings from the Phase 1 iteration loop, not entered by
Phase 2 or the scanner fix:

- **Service-shutdown timeout** — `nucleus.Run`'s `wg.Wait()` after
  `cancelServices()` has no deadline.
- **`Lifecycle.OnShutdown` context deadline** — derived from
  `context.Background()` with no bound.
- **`joinPath` double-slash collapse** — `routerAdapter.joinPath`
  produces `/x/x/123` when `prefix=/x` and `p=/x/123`.

## Files of interest

- `docs/iterations/2026-05-20-contract-freeze-scanner-constructors.md`
  — this session's archive.
- `docs/iterations/2026-05-20-adr010-phase2bcd-config-loader-completion.md`
  — the ADR-010 §2 completion archive.
- `contracts/freeze_test.go` — scanner; the `packages` slice is the
  target for candidate #2 (package-coverage gap).
- `pkg/admin/` — target for candidate #1.
- `docs/adrs/ADR-010-fluent-api-v2-pkg-nucleus.md` — Phase 3 / Phase 4
  remain.

## Notes / decisions log

- 2026-05-20 — Freeze-scanner constructor-gap fix shipped (PR #77).
  Pure governance tooling; no CHANGELOG, no semver bump, no ADR.
  Opened candidate #2 (the package-coverage gap) as the natural
  follow-up.
- 2026-05-20 — ADR-010 §2 four-phase slicing complete (Phases
  2b/2c/2d). §16 clarified (namespacing on both tracking tables).
- 2026-05-16 — Owner sliced ADR-010 §2 into 2a/2b/2c/2d.
