# Current Iteration

> Owned by `session-curator`. Edited by other subagents only via the
> Session Start / Session End protocols (`CLAUDE.md` §2 and §5).

## Goal

No active iteration. Last completed: **`pkg/admin` bootstrap users-table
dialect-aware DDL fix** (PR #78 → `2975108`), archived at
`docs/iterations/2026-05-20-admin-bootstrap-ddl-dialect-aware.md`. The
admin bootstrap now works on MSSQL and Oracle; a chained Oracle scaffold
`/`-terminator bug was fixed in the same PR. Two deeper Oracle issues
surfaced and were de-scoped to candidates #2 / #3 below.

## Scope

- in: (TBD — owner to confirm scope when iteration is registered)
- out: (TBD)

## Acceptance criteria

- [ ] (TBD)

## Status

### Done (2026-05-20)

- **Admin bootstrap DDL dialect-aware fix** (PR #78 → `2975108`).
  Closes the candidate-#1 bug (MSSQL `Incorrect syntax`, Oracle
  `ORA-03076`). Also fixed the chained Oracle scaffold `/`-terminator
  bug (`ORA-06550`). MSSQL `TestSQLMatrix_AutoMigrate_Exploratory`
  re-wired (green); Oracle lane deferred (candidate #2). Semver: patch.
- **Freeze-scanner constructor-gap fix** (PR #77 → `28f75b2`).
- **ADR-010 §2 config loader feature-complete** (Phases 2a–2d, PRs
  #73–#76).

### Done (earlier — see prior archives)

- v0.7.0 (PRs #56–#59); CSRF hardening (ADR-006); slog redaction
  (ADR-007); CSRF follow-ups + schema drift (ADR-008 + ADR-009);
  MSSQL/Oracle SchemaDrift (#66); pkg/app+pkg/nucleus inventory (#65);
  ADR-010 Phase 1 + examples purge (#71).

### In progress

- (none)

### Blocked

- (none)

## Candidate next steps (priority order, pending owner confirmation)

1. **Freeze-scanner package-coverage gap.** `contracts/freeze_test.go`'s
   `packages` slice omits six `pkg/*` packages (`admin`, `circuit`,
   `health`, `observability`, `openapi`, `outbox`) — zero
   removal-protection on their exported surface. Audit each; confirm
   lifecycle posture first (`pkg/outbox.NewKafkaBridge` must NOT be
   frozen until Kafka lands; `pkg/openapi` is experimental;
   `pkg/admin`/`pkg/outbox` are transitional). A code comment at the
   slice documents the deliberate omission.

2. **Oracle model-scaffold identifier-casing (opened by PR #78).**
   `BuildOracleMigrationScaffold` quotes identifiers
   (`CREATE TABLE "ci_automig_live_users"` → case-sensitive lowercase),
   diverging from the unquoted-uppercase convention the rest of the
   Oracle path uses and `USER_TAB_COLUMNS` introspection expects. Blocks
   the Oracle `TestSQLMatrix_AutoMigrate_Exploratory` lane (deferred
   with a NOTE breadcrumb in `.github/workflows/ci.yml`). Needs a
   decision on the framework's Oracle identifier strategy
   (quoted-lowercase vs. unquoted-uppercase) incl. reserved-word and
   query/CRUD-layer implications — likely an ADR. When it lands, re-add
   the Oracle AutoMigrate_Exploratory test line.

3. **Oracle multi-block AutoMigrate execution (opened by PR #78).**
   Scaffolds for models with secondary indexes emit multiple
   `BEGIN…END;` PL/SQL blocks; the single-`Exec` AutoMigrate path (and
   the file Migrator's `tx.Exec`) can't run them as one batch. Needs a
   statement-splitting executor.

4. **ADR-010 Phase 3 — `/_/config` + `nucleus config print
   --effective`.** Compliance items #6, #12, #13. Auth-gated by
   `WithAdmin()` (Casbin default-deny); redaction via
   `observe.DefaultRedactedKeys()`. Requires per-key source tracking the
   Phase 2 loader does not yet capture.

5. **`session_cookie_secure` default `false`** (Phase 2b security-
   auditor MED-1). Pre-existing security default; the non-nullable
   mechanism doesn't cover it (default already permissive). Flip to
   `true` or add to the non-nullable set.

6. **ADR-010 §2 layer 3 — field-semantic validation** (ranges, enums,
   parseable durations; ADR-010 §96 layer 3). Standalone follow-up on
   the now-complete merge engine.

7. **ADR-010 Phase 4 — Docs-sync + website + new reference applications
   under a freshly-scoped `examples/`.** Target: v0.9.X.

8. **Cloud Secrets Provider plugin extraction (AWS → GCP → Azure →
   Vault).** Removes AWS SDK from core `go.mod`.

9. **Column-type comparison in `SchemaDrift`.** Cross-dialect
   type-family compatibility table.

10. **SchemaDrift end-to-end usage guide** in
    `docs/guides/MODELING_MULTI_DATABASE.md`.

11. **`go mod tidy` unblock** (admin/proto replace-directive).

12. **`tasks.Manager` struct→interface DEP** (optional DEP-2026-004).

13. **Audit §7 menores** — 503 path test for `/healthz`,
    endpoints-parity doc-parsing, `pkg/health/{db,redis,storage}.go`
    tests.

## Carry-forward follow-ups (ADR-010 Phase 1, still open)

- **Service-shutdown timeout** — `nucleus.Run`'s `wg.Wait()` after
  `cancelServices()` has no deadline.
- **`Lifecycle.OnShutdown` context deadline** — derived from
  `context.Background()` with no bound.
- **`joinPath` double-slash collapse** — `routerAdapter.joinPath`
  produces `/x/x/123` when `prefix=/x` and `p=/x/123`.

## Files of interest

- `docs/iterations/2026-05-20-admin-bootstrap-ddl-dialect-aware.md` —
  this session's archive (full bug-chain record).
- `pkg/model/migration_scaffold_oracle.go` — candidate #2 target
  (identifier quoting).
- `.github/workflows/ci.yml` — Oracle AutoMigrate_Exploratory NOTE
  breadcrumb (re-add the line when candidate #2 lands).
- `contracts/freeze_test.go` — candidate #1 target (package-coverage gap).
- `pkg/nucleus/config.go`, `pkg/nucleus/nucleus.go` — Phase 2 loader.

## Notes / decisions log

- 2026-05-20 — PR #78 (admin bootstrap DDL + Oracle scaffold `/`).
  Discovered a chain of 4 latent Oracle bugs (admin DDL masked them);
  fixed 2, de-scoped 2 (#2 identifier-casing, #3 multi-block exec) as
  their own candidates. Kept the Oracle AutoMigrate_Exploratory lane
  deferred — disciplined scoping over rabbit-holing the identifier
  decision.
- 2026-05-20 — Freeze-scanner constructor-gap fix (PR #77); ADR-010 §2
  complete (Phases 2b/2c/2d).
