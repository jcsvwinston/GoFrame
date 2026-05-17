# Current Iteration

> Owned by `session-curator`. Edited by other subagents only via the
> Session Start / Session End protocols (`CLAUDE.md` §2 and §5).

## Goal

No active iteration. Two iterations completed on 2026-05-16:

- **ADR-010 Phase 1 Foundation + wholesale `examples/*` removal**
  (PR #71 → `cdc0a76`) — archived at
  `docs/iterations/2026-05-16-adr010-phase1-and-examples-purge.md`.
- **ADR-010 Phase 2a — `FromConfigFile` single-file loader**
  (PR #73 → `2b650f3`) — archived at
  `docs/iterations/2026-05-16-adr010-phase2a-fromconfigfile-single-file.md`.

Owner sliced ADR-010 §2 ("Config loading + merge engine") into four
sub-iterations **2a / 2b / 2c / 2d**, of which 2a is now done. Next
ADR-010 sub-iteration is **Phase 2b** (multi-file merge with
`_append` / `_remove` suffix operators + TOML / JSON parsers +
non-nullable security keys). Candidate #1 (`pkg/admin` MSSQL/Oracle
bootstrap DDL fix) remains the top-ranked non-Phase-2 alternative if
owner wants to interleave.

## Scope

- in: (TBD — owner to confirm scope when iteration is registered)
- out: (TBD)

## Acceptance criteria

- [ ] (TBD)

## Status

### Done (2026-05-16)

- **ADR-010 Phase 2a — `FromConfigFile` single-file loader**
  (PR #73, merged as `2b650f3`). Replaces the Phase 1
  `ErrConfigLoaderNotImplemented` stub with a real YAML loader:
  1 MiB per-file size cap via `io.LimitReader` (DoS-safe against
  anchor-expansion in `gopkg.in/yaml.v3`); extension-based parser
  inference (`.yaml`/`.yml` work; `.toml`/`.json` return Phase 2b
  sentinel); strict-unknown-keys schema validation against
  `app.ContractConfigKeyPatterns()` with Levenshtein-≤3 did-you-mean
  hints; wildcard pattern matching for `databases.*.url`-style
  schema slots. New exported names: `MaxConfigFileBytes`,
  `ErrConfigFileTooLarge`, `ErrUnsupportedConfigFormat`,
  `ErrUnknownConfigKeys`. Freeze baseline rebased (+4 / −1).
  `pkg/nucleus/config_test.go` covers 13 cases.
- **ADR-010 Phase 1 Foundation + wholesale `examples/*` removal**
  (PR #71, merged as `cdc0a76`). `pkg/nucleus` rewritten with the
  canonical `App{}` struct, `AppBuilder` fluent chain, `Router`
  interface, `Module[C any]` generic, three-surface equivalence test,
  `Patch` flat-declarative method, and freeze-scanner baseline
  reseeded with 101 entries. Every `examples/*` tree deleted, runnable
  lab scripts removed, `Dockerfile` builds `./cmd/nucleus`,
  compatibility harness reduced to the `core-build` placeholder.
  ADR-010 Status flipped Proposed → Accepted. CHANGELOG carries two
  `BREAKING (...)` entries. All 11 CI checks SUCCESS (including the
  four live-DB lanes); semver bump hint: minor (`v0.8.0`).
- v0.7.0 released (PRs #56–#59); CSRF hardening (ADR-006, PR #60);
  slog secret redaction (ADR-007, PR #62); CSRF follow-ups + schema
  drift (ADR-008 + ADR-009, PR #63); MSSQL/Oracle SchemaDrift (PR #66);
  pkg/app + pkg/nucleus inventory (PR #65) — see prior archived
  iterations.

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
   `.github/workflows/ci.yml`.

2. **ADR-010 Phase 2b — multi-file merge engine + TOML/JSON parsers
   + non-nullable security keys.** ADR-010 §3 merge semantics with
   `_append` / `_remove` suffix operators; TOML and JSON file
   parsers (currently return `ErrUnsupportedConfigFormat`);
   non-nullable security keys (`cors.origins`, `auth.providers`,
   `authz.policy_path`, `session.secret`) per ADR-010 §14;
   `nucleus.WithConfigStrict(true)` for mixed-format rejection.
   This is the **next ADR-010 sub-iteration**; Phase 2a closed the
   single-file path.

3. **ADR-010 Phase 2c — production-strict guard.**
   `nucleus.WithUnknownFields("warn")` opt-out from the strict
   schema validation; `NUCLEUS_ENV=production` env var that overrides
   to strict regardless of code-level setting; the startup `WARN`
   line per ADR-010 §15.

4. **ADR-010 Phase 2d — module migration namespacing.** Update
   `pkg/db/migrate.go`'s `migrationsChecksumsTable` so migration file
   checksums are keyed as `<module_name>/<filename>` rather than
   `<filename>` per ADR-010 §16. Prevents cross-module filename
   collisions when multiple modules share a database alias.

5. **ADR-010 Phase 3 — `/_/config` + `nucleus config print
   --effective`.** Compliance items #6, #12, #13. Auth-gated by
   `WithAdmin()` (Casbin default-deny); redaction via
   `observe.DefaultRedactedKeys()`.

6. **ADR-010 Phase 4 — Docs-sync + website + new reference
   applications under a freshly-scoped `examples/`.** Target: v0.9.X.
   This is where the replacement examples get authored, the website
   docs are rewritten, and the manifest pattern is introduced. The
   compatibility harness fixture profiles (`minimal-api`,
   `admin-heavy`, `plugin-heavy`) return here.

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
    elegantly, are gone entirely once candidate #7 lands).

11. **`tasks.Manager` struct→interface DEP** — optional DEP-2026-004
    for the binary-incompatible type-identity change.

12. **Audit §7 menores** — 503 path test for `/healthz`,
    endpoints-parity doc-parsing, individual tests for
    `pkg/health/{db,redis,storage}.go`.

## Carry-forward follow-ups (annotated by Phase 1 subagent loop)

These are non-blocker findings from the Phase 1 iteration loop. Each
belongs in ADR-010 Phase 2 when the relevant code path is reworked:

- **Service-shutdown timeout** — `nucleus.Run`'s `wg.Wait()` after
  `cancelServices()` has no deadline. A misbehaving service that
  ignores ctx cancellation blocks `Lifecycle.OnShutdown` indefinitely.
  Use `withTimeoutFromConfig`-derived deadline.
- **`Lifecycle.OnShutdown` context deadline** — same family of issue;
  the context is currently derived from `context.Background()` with no
  bound.
- **`joinPath` double-slash collapse** — `routerAdapter.joinPath`
  produces `/x/x/123` when `prefix=/x` and `p=/x/123`. User error, but
  defensive `path.Join`-style handling avoids the footgun once nested
  `Group` calls are common.

## Files of interest

- `docs/iterations/2026-05-16-adr010-phase1-and-examples-purge.md` —
  archived ADR-010 Phase 1 iteration.
- `docs/iterations/2026-05-16-adr010-phase2a-fromconfigfile-single-file.md`
  — archived ADR-010 Phase 2a iteration.
- `docs/adrs/ADR-010-fluent-api-v2-pkg-nucleus.md` — Phase 2b /
  Phase 2c / Phase 2d / Phase 3 / Phase 4 are the remaining
  ADR-010 phases.
- `pkg/admin/` — target for candidate #1 (admin bootstrap DDL).
- `pkg/nucleus/config.go` — Phase 2a loader; Phase 2b adds the merge
  engine and the TOML/JSON parsers on top.

## Notes / decisions log

- 2026-05-16 — ADR-010 Phase 1 + `examples/*` purge completed and
  merged as PR #71 (`cdc0a76`); archived at
  `docs/iterations/2026-05-16-adr010-phase1-and-examples-purge.md`.
- 2026-05-16 — Owner sliced ADR-010 §2 ("Config loading + merge
  engine") into four sub-iterations **2a / 2b / 2c / 2d** so each
  lands as its own reviewable PR: 2a = single-file loader + size cap
  + schema validation; 2b = multi-file merge + `_append`/`_remove` +
  TOML/JSON + non-nullable security keys; 2c =
  `WithUnknownFields("warn")` + `NUCLEUS_ENV=production` strict
  override + startup WARN; 2d = module migration namespacing in
  `pkg/db/migrate.go`.
- 2026-05-16 — ADR-010 Phase 2a (single-file `FromConfigFile` loader)
  completed and merged as PR #73 (`2b650f3`); archived at
  `docs/iterations/2026-05-16-adr010-phase2a-fromconfigfile-single-file.md`.
- 2026-05-17 — Both Phase 1 and Phase 2a state-close work batched
  into a single `/handoff` commit. The owner deferred the Phase 1
  state-close at the end of the 2026-05-16 session; Phase 2a shipped
  the same day deliberately leaving state files untouched (per the
  convention established by #61 / #64 / #68); this `/handoff` closes
  both iterations together.
