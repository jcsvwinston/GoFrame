# Handoff — last session closing note

> Owned by `session-curator`. Overwritten at the end of every session
> by `/handoff`. Read first by `/resume` at the start of the next one.

ITERATION:    `pkg/admin` bootstrap users-table dialect-aware DDL fix — COMPLETE (PR #78 → `2975108`), archived at `docs/iterations/2026-05-20-admin-bootstrap-ddl-dialect-aware.md`. No active iteration.
BRANCH:       main (in sync with origin/main, 0 ahead / 0 behind). Feature branch merged + deleted.
LAST COMMIT:  2975108 fix(admin): dialect-aware bootstrap users-table DDL for MSSQL/Oracle (#78)
STATUS:       Admin bootstrap now works on MSSQL + Oracle (dialect-aware CREATE TABLE: mssql IF-OBJECT_ID/NVARCHAR/BIT, oracle PL/SQL block + VARCHAR2/NUMBER(1) DEFAULT-before-NOT-NULL). Chained Oracle scaffold `/`-terminator bug (ORA-06550) fixed in the same PR. `BootstrapAdminConfig` gained an additive `System` field. MSSQL `TestSQLMatrix_AutoMigrate_Exploratory` re-wired (green); Oracle lane deferred. CI 9/9 SUCCESS incl. all four live-DB lanes. Semver: patch.
NEXT STEP:    Owner picks the next iteration from `CURRENT_ITERATION.md` §"Candidate next steps". Top picks: (1) freeze-scanner package-coverage gap (6 omitted pkg/* packages); (2) Oracle model-scaffold identifier-casing (unblocks the deferred Oracle AutoMigrate lane — likely an ADR); (3) ADR-010 Phase 3 (`/_/config`).
BLOCKERS:     none.
FILES OF INTEREST:
  - .claude/state/CURRENT_ITERATION.md — reset to "no active iteration"; 13 candidates + 3 Phase-1 carry-forwards. Candidates #2/#3 are the new Oracle follow-ups opened by PR #78.
  - docs/iterations/2026-05-20-admin-bootstrap-ddl-dialect-aware.md — this session's archive (the full Oracle bug-chain story).
  - pkg/model/migration_scaffold_oracle.go — candidate #2 target (identifier quoting).
  - .github/workflows/ci.yml — Oracle AutoMigrate_Exploratory NOTE breadcrumb marks where to re-add the test line once candidate #2 lands.
  - contracts/freeze_test.go — candidate #1 target (package-coverage gap).

NOTES:
  - **Oracle bug chain (this session):** the admin-bootstrap DDL bug ran during App.New, masking downstream bugs. Fixing it exposed: (2) the scaffold `/` terminator (fixed), then (3) the scaffold identifier-casing mismatch (de-scoped — `BuildOracleMigrationScaffold` quotes lowercase, the rest of the framework + USER_TAB_COLUMNS introspection expect unquoted-uppercase), then (4) multi-block AutoMigrate execution (de-scoped). #3 and #4 are candidates #2 and #3 in the queue. The disciplined call was to ship the two validated fixes and NOT rabbit-hole the Oracle identifier-strategy decision into this PR.
  - **Re-adding the Oracle AutoMigrate lane:** once candidate #2 (identifier-casing) lands, re-add `go test -tags oracle ./pkg/app -run '^TestSQLMatrix_AutoMigrate_Exploratory$' -v` to the oracle lane in `.github/workflows/ci.yml` (the NOTE breadcrumb marks the spot). The MSSQL equivalent is already wired and green.
  - **Pre-existing, out-of-scope (noted by security-auditor):** the admin bootstrap INSERT uses `fmt.Sprintf` + `quoteBootstrapSQLString` rather than driver placeholders. Operator config + bcrypt-hash values only, single call site, constant table name — low risk; a parameterised-query hardening pass is a separate follow-up.
  - State-close convention: feature PRs #77/#78 left `.claude/state/*` untouched; this `/handoff` is the state-close. Per #61/#64/#68/#70/f1453b5/76c9c95.

OPEN HOUSEKEEPING (carried, none blocking):
  - `go mod tidy` cannot run cleanly (admin/proto replace-directive). The Phase 2b/2c koanf sub-modules (toml/v2, json, confmap) are stuck as `// indirect`. Moot once the Cloud Secrets plugin extraction lands.
  - `panic(` count in non-test code reportedly 4→0 since b1e497e — unconfirmed. NewModuleMigrator (Phase 2d) adds 2 deliberate MustCompile-style constructor panics; intentional.

Updated: 2026-05-20
