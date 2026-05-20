# Iteration archive — 2026-05-20 admin bootstrap DDL dialect-aware fix (+ Oracle scaffold `/` fix)

> Archived 2026-05-20 as part of the session-end `/handoff`. Shipped as
> PR #78 (`2975108`). Candidate #1 from the prior queue. The iteration
> uncovered and resolved a chain of latent bugs; two deeper Oracle
> issues were deliberately de-scoped to their own follow-ups.

## Goal

Fix the `pkg/admin` bootstrap users-table DDL so it works on MSSQL and
Oracle (candidate #1), and re-enable the previously-disabled exploratory
test `TestSQLMatrix_AutoMigrate_Exploratory` that the bug had blocked
since PR #66.

## The bug chain

The admin-bootstrap DDL bug ran during `App.New` (gated by
`AdminBootstrapEmail`, before AutoMigrate), so it masked the bugs
downstream of it. Fixing each one exposed the next:

1. **Admin bootstrap DDL not dialect-aware** (the candidate).
   `admin.EnsureBootstrapAdminUser` emitted one hardcoded
   `CREATE TABLE IF NOT EXISTS … INTEGER NOT NULL DEFAULT 0 … TEXT`
   form. MSSQL rejected it (`Incorrect syntax near 'nucleus_admin_users'`
   — no `IF NOT EXISTS`), Oracle rejected it (`ORA-03076` — Oracle
   requires `DEFAULT <v> NOT NULL`, not the reverse).
2. **Oracle model scaffold `/` terminator.** Once #1 was fixed, the
   Oracle lane reached AutoMigrate and failed with `ORA-06550`:
   `pkg/model.BuildOracleMigrationScaffold` (via `writeOraclePLSQLBlock`)
   emitted a SQL\*Plus `/` terminator after each PL/SQL block — invalid
   PL/SQL when sent straight to go-ora (the AutoMigrate `sqlDB.Exec` and
   the file-Migrator `tx.Exec` paths never go through SQL\*Plus).
3. **Oracle scaffold identifier-casing (DE-SCOPED).** Once #2 was fixed,
   AutoMigrate ran without error but the test's introspection found no
   columns: `BuildOracleMigrationScaffold` QUOTES identifiers
   (`CREATE TABLE "ci_automig_live_users"` → case-sensitive lowercase),
   while the rest of the framework's Oracle path (pkg/db migrations,
   admin bootstrap) uses UNQUOTED identifiers folded to UPPERCASE — what
   `USER_TAB_COLUMNS` introspection expects.

## Status

### Done (2026-05-20, PR #78 → `2975108`)

- **Admin bootstrap DDL is dialect-aware** (`pkg/admin/bootstrap_admin.go`).
  New `bootstrapAdminUsersTableDDL(system)` switch mirroring
  `pkg/db`'s `migrationsTableDDL`: mssql `IF OBJECT_ID … NVARCHAR/BIT`;
  oracle `BEGIN EXECUTE IMMEDIATE … VARCHAR2 / NUMBER(1) DEFAULT 0 NOT
  NULL` (DEFAULT before NOT NULL — the ORA-03076 fix); default
  unchanged. `BootstrapAdminConfig` gains an additive `System string`
  field; `pkg/app/app.go` passes `adminAuthDB.System()`. Empty System →
  portable DDL (backward compatible). MSSQL + Oracle admin bootstrap
  now succeed in CI.
- **Oracle model scaffold `/` terminator removed**
  (`pkg/model/migration_scaffold_oracle.go::writeOraclePLSQLBlock`).
  Blocks now end with bare `END;\n`, matching the no-`/` PL/SQL blocks
  pkg/db already uses. Oracle `AutoMigrate` executes without error.
  Regression test asserts no standalone `/` line + balanced BEGIN/END;.
- **`TestSQLMatrix_AutoMigrate_Exploratory` re-wired into the MSSQL CI
  lane** (fully green). The Oracle lane is deferred with a precise NOTE
  breadcrumb in `.github/workflows/ci.yml` pointing at follow-up #3.
- Tests: `pkg/admin/bootstrap_admin_test.go` (new) — per-dialect DDL
  assertions (mssql IF-OBJECT_ID guard, oracle DEFAULT-before-NOT-NULL
  ORA-03076 guard, dialect column types), sqlite end-to-end bootstrap,
  empty-System fallback. `pkg/model` regression test for the `/`
  removal.
- CI: 9/9 checks SUCCESS including all four live-DB lanes. Semver:
  patch (bug fix; additive `System` field; no symbol removed/renamed).

### Iteration loop

Admin-DDL change: full 9-subagent loop, all green/PASS/NITS, no BLOCKER
(architect PASS — dialect conventions match pkg/db/pkg/model; security
PASS — no injection surface, INSERT-quoting note pre-existing/out-of-
scope; contract PASS — pkg/admin not in freeze scan, `System` additive).
Oracle scaffold `/` delta: focused code-reviewer pass, PASS.

### Scope discipline

Bug #3 (Oracle identifier-casing) is a genuine architectural decision —
the framework's Oracle identifier strategy (quoted-lowercase vs.
unquoted-uppercase, with reserved-word and query-layer implications).
Rather than rabbit-hole it into this PR, it was de-scoped to its own
iteration; the two validated fixes shipped and the Oracle AutoMigrate
lane stays deferred with a breadcrumb.

### Blocked
- (none)

## Follow-ups opened by this iteration

1. **Oracle model-scaffold identifier-casing (candidate #3).**
   `BuildOracleMigrationScaffold` quotes identifiers, diverging from the
   unquoted-uppercase convention used everywhere else in the Oracle
   path and expected by `USER_TAB_COLUMNS` introspection. Blocks the
   Oracle `TestSQLMatrix_AutoMigrate_Exploratory` lane. Needs a decision
   on the framework's Oracle identifier strategy (likely an ADR — it
   affects scaffolds, SchemaDrift, and the query/CRUD layer). When it
   lands, re-add the Oracle `go test … AutoMigrate_Exploratory` line
   (the NOTE breadcrumb in `.github/workflows/ci.yml` marks the spot).
2. **Oracle multi-block AutoMigrate execution (candidate #4).**
   Scaffolds for models with secondary indexes emit multiple
   `BEGIN…END;` blocks; the single-`Exec` AutoMigrate path can't run
   them as one batch. Affects file-based migration too. A
   statement-splitting executor is the likely fix.

## Files of interest

- `pkg/admin/bootstrap_admin.go` — dialect-aware bootstrap DDL.
- `pkg/model/migration_scaffold_oracle.go` — `/` removed; candidate #3
  target (identifier quoting).
- `.github/workflows/ci.yml` — MSSQL AutoMigrate_Exploratory wired;
  Oracle deferred with breadcrumb.
- `pkg/app/app.go` — passes `adminAuthDB.System()` to the bootstrap.

## Notes / decisions log

- 2026-05-20 — Admin bootstrap DDL fix shipped (PR #78). MSSQL +
  Oracle bootstrap green. Discovered the Oracle scaffold `/` bug as a
  consequence and fixed it in the same PR. De-scoped the Oracle
  identifier-casing bug (#3) and the multi-block execution gap (#4) to
  their own follow-ups; kept the Oracle AutoMigrate_Exploratory lane
  deferred with a breadcrumb. Semver: patch.
