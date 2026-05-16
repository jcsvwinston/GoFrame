# Current Iteration

> Owned by `session-curator`. Edited by other subagents only via the
> Session Start / Session End protocols (`CLAUDE.md` §2 and §5).

## Goal

**ADR-010 Phase 1 — Foundation of the Fluent API v2 for `pkg/nucleus`,
delivered together with a clean-break removal of every `examples/*` tree.**

Phase 1 pins the new public shape of `pkg/nucleus` (canonical `App{}`
struct, `Module[C any]` + `ModuleSpec`, `Router` interface, three-surface
equivalence test, freeze-scanner seed). ADR `Status` flips from
`Proposed` to `Accepted` in this PR following the ADR-006 / ADR-008
acceptance pattern.

The original ADR-010 §263 / §297 stipulated rewriting the two
`examples/ecommerce_dashboard/backend/*` consumers in the same PR.
**Owner decision (2026-05-16):** instead of rewriting any consumer,
**every `examples/*` tree is deleted** in this same iteration. The
existing examples are obsolete relative to the framework's current
shape; rebuilding them now would only add noise. New, solid reference
applications will be authored in v0.9.X once Phases 2–4 (config loader,
`/_/config`, docs-sync) have landed and the surface is stable.

This decision is consistent with the pre-`v1.0`, single-maintainer,
no-external-users posture documented in `docs/governance/COMPATIBILITY_SLO.md`
and the ADR-006 / ADR-008 clean-break precedent. No DEP or MA artefacts
are produced.

## Scope

### In

1. **`pkg/nucleus` Phase 1 rewrite** per ADR-010 §1, §6, §7 — Compliance
   items #1, #2, #3, #5, #11:
   - `pkg/nucleus/nucleus.go` rewritten to expose the new `AppBuilder`
     chain (`FromConfigFile` *stub*, `Use`, `Mount`, `WithoutDefaults`,
     `WithExtensions`, `Build`, `Start`/`Serve`) and the canonical
     `App{}` struct embedding `app.Config` plus the four Go-only wiring
     fields (`Modules`, `Middleware`, `Services`, `Lifecycle`).
   - `pkg/nucleus/router.go` (new) — `Router` interface with `Get`,
     `Post`, `Put`, `Delete`, `Group`, `Resource(...)` + explicit
     `Methods(...)` variadic. No reflection-based method discovery.
   - `pkg/nucleus/module.go` (new) — type-erased `ModuleSpec` interface
     and generic `Module[C any]` constructor with `Build() ModuleSpec`.
   - `pkg/nucleus/equivalence_test.go` (new) — three-surface equivalence
     test per ADR-010 §1 normalisation rules (sorted map keys, function
     reference identity).
   - `FromConfigFile` lands as a Phase 2 stub for now (returns
     `ErrNotImplemented` or equivalent) — its five-layer validator is
     Phase 2 work. The signature is in place so the builder shape is
     complete and the equivalence test compiles.
   - Freeze-scanner seed at the end of Phase 1:
     `NUCLEUS_UPDATE_CONTRACT_BASELINE=1 go test ./contracts/...` to
     populate `contracts/baseline/api_exported_symbols.txt` for the new
     `pkg/nucleus` surface (closes the false-green flagged in the
     2026-05-15 inventory).

2. **Wholesale deletion of `examples/`** — replaces the original
   "rewrite two consumers" path in ADR-010 §263, §284, §297:
   - Delete: `examples/admin-quickstart/`, `examples/balancer/`,
     `examples/ecommerce_dashboard/`, `examples/fleetmanager/`,
     `examples/ministore/`, `examples/mvc_api/`, `examples/plugins/`
     (and the parent `examples/` directory once empty).
   - **Scrub every reference to `examples/*` in shipped files:**
     - `.github/workflows/ci.yml` (lines 53, 56 — race lane `go test`
       targets and the admin-smoke step).
     - `.github/workflows/release.yml` (line 35 — release smoke step).
     - `scripts/ci/run_compatibility_harness.sh` (lines 94–96 — drop
       the `minimal-api`, `admin-heavy`, `plugin-heavy` profiles; they
       are integration fixtures backed by `examples/` and have nowhere
       to point during this window).
     - `scripts/release/rehearse_rc.sh` (line 34).
     - `scripts/release/generate_compatibility_report.sh` (line 103 —
       `stable-plugin-sdk` check).
     - `scripts/cluster-start.sh` (line 14 — `go build … ./examples/mvc_api`).
     - `scripts/dev/run_admin_cluster_lab.sh` (lines 182, 235) and the
       matching `.ps1` if it carries the same.
     - `README.md` (the example table at lines ~187–192).
     - `docs/ADMIN_CLUSTER_LAB.md`, `docs/guides/MAIL_GUIDE.md`,
       `docs/reference/DEVELOPER_MANUAL.md`,
       `docs/reference/PLUGIN_SDK.md`,
       `docs/reference/PLUGIN_EXAMPLES.md`,
       `docs/governance/ENTERPRISE_LONG_TERM_ROADMAP.md`,
       `docs/adrs/ADR-004-casbin-default-deny-mount.md`,
       `docs/deprecations/DEP-2026-002-builtin-sendgrid-provider.md`,
       `docs/migration_assistants/MA-2026-002-sendgrid-builtin-to-plugin.md`.
       Each reference is either removed or rewritten as "previously
       lived under `examples/<name>` — to be restored in v0.9.X with
       the new reference applications" depending on whether the
       surrounding sentence still makes sense without it.
   - **Leave references intact** in: `docs/iterations/*` and
     `docs/reports/*` (historical archives — rewriting history is a
     non-goal); `docs/audits/2026-05-12-enterprise-readiness.md`
     (audit snapshot, also historical).
   - The `docs/audits/` snapshot reference can be left in place because
     it is dated and labelled as a point-in-time observation.

3. **ADR-010 update in the same PR:**
   - Flip `Status: Proposed` → `Status: Accepted` and add a Date line
     for acceptance (per ADR-006 / ADR-008 pattern).
   - Revise §263 (Negative consequences) to drop the "two consumer
     files rewritten" language and replace with the
     deletion-then-rebuild-in-v0.9.X framing.
   - Revise §244 (Compliance #4 in §9 Documentation-synchronisation):
     "Code blocks in website docs are imported from `examples/*` via
     Docusaurus include syntax" is deferred to v0.9.X. The Phase 4
     docs-sync iteration will pick this back up when there are new
     examples to import from. Reword to note the deferral; do not
     delete the commitment.
   - Revise Compliance item #1 to drop the "two consumer files updated
     in same PR" clause and replace with "all `examples/*` removed in
     same PR; new reference applications authored in v0.9.X".
   - Revise §270 (Negative consequences, "Website needs a substantial
     rewrite") — note that the website rewrite *and* the new
     reference applications are now bundled into Phase 4 / v0.9.X.

4. **CHANGELOG.md `Unreleased`** — two `### Changed` entries with
   inline `BREAKING (...)` labels following ADR-006 / ADR-008 form:
   - `BREAKING (pkg/nucleus rewrite): legacy fluent chain replaced
     with AppBuilder + canonical App{} struct per ADR-010 Phase 1.`
   - `BREAKING (examples/* removed): every example application removed;
     new reference applications will land in v0.9.X.`

### Out (deferred to later iterations)

- ADR-010 Phase 2: five-layer validator and merge engine with
  `_append`/`_remove` suffix operators (Compliance #4, #7, #14, #15,
  #16, #17).
- ADR-010 Phase 3: `/_/config` endpoint and `nucleus config print
  --effective` (Compliance #6, #12, #13).
- ADR-010 Phase 4: docs-sync mechanism, website rewrite, and **the new
  reference applications under a freshly-scoped `examples/`** (Compliance
  #8 — already landed via PR #67; #9, #10). Target: v0.9.X.
- `pkg/admin` MSSQL/Oracle bootstrap DDL fix (was top-ranked candidate;
  bumped one slot — re-promote after Phase 1 lands).
- Cloud Secrets Provider plugin extraction (AWS).

## Acceptance criteria

- [ ] `pkg/nucleus/nucleus.go` exposes the new `AppBuilder` chain and
      canonical `App{}` struct. Legacy fluent methods (`Port`, `Host`,
      `SQLite`, `Postgres`, `MySQL`, `WithAdmin`, `SPA`, `Templates`,
      `Static`, `Cors`, `Provide`, `Model`, `AutoMigrate`, `Run`) are
      gone — no shims, no WARN wrappers (pre-`v1.0`, per ADR-006 /
      ADR-008 precedent).
- [ ] `pkg/nucleus/router.go` defines the `Router` interface with the
      seven methods enumerated in ADR-010 §195–§204 and `Resource(...)`
      requires an explicit `nucleus.Methods(...)` variadic.
- [ ] `pkg/nucleus/module.go` defines `ModuleSpec` (type-erased) and
      `Module[C any]` (generic) with `Build() ModuleSpec`.
- [ ] `pkg/nucleus/equivalence_test.go` passes — all three surfaces
      (fluent, direct-struct, bootstrap pattern) produce equal
      `nucleus.App{}` values under the normalisation rules in
      ADR-010 §79–§85.
- [ ] `go test -race ./pkg/... ./internal/cli ./cmd/nucleus` green.
- [ ] `bash scripts/ci/check_contract_freeze.sh` green after the
      `pkg/nucleus` baseline reseed; `contracts/baseline/api_exported_symbols.txt`
      now contains real `pkg/nucleus` entries (no longer false-green).
- [ ] No file under `examples/` exists in the working tree.
- [ ] No file outside `docs/iterations/`, `docs/reports/`,
      `docs/audits/` references the literal path `examples/` (verify
      with `git grep -nE 'examples/[a-z]'`).
- [ ] `.github/workflows/{ci,release}.yml` and `scripts/{ci,release,dev}/*`
      no longer reference any `examples/*` path.
- [ ] `scripts/ci/run_compatibility_harness.sh` no longer carries the
      `minimal-api`, `admin-heavy`, `plugin-heavy` profiles. Either the
      profiles are removed entirely or the harness is reduced to its
      remaining (non-examples) profiles — the doc comment in the script
      explains the temporary reduction and links back to the v0.9.X
      restoration.
- [ ] `README.md` no longer lists examples; the section is replaced
      with a short "Reference applications will be authored in v0.9.X"
      note pointing at ADR-010.
- [ ] `docs/adrs/ADR-010-fluent-api-v2-pkg-nucleus.md`: Status is
      `Accepted`; §263, §244 (Compliance §9 #4), §297, Compliance #1
      reflect the deletion-not-rewrite decision.
- [ ] `CHANGELOG.md` `Unreleased` block carries the two `BREAKING (...)`
      entries described above.
- [ ] Iteration loop steps 1–9 all green (`/iterate` orchestration).

## Status

### Done

- (none yet — iteration registered 2026-05-16, work has not started)

### In progress

- (none yet)

### Blocked

- (none)

## Subagent guidance for this iteration

This iteration deliberately inverts or narrows the scope of two
subagents; explicit guidance is recorded here so `/iterate` does not
have to re-discover it on every loop.

- **`architect-reviewer`** — required. Validate against ADR-010 §1
  (canonical struct + three surfaces), §6 (`Module` contract), §7
  (Router three styles). Confirm Phase 1 boundary is honoured: no
  Phase 2 validator code, no `/_/config` work, no docs-sync work.
  Confirm ADR-010 textual revisions are coherent with the §1–§9
  decision body (the consequence and compliance sections must reflect
  the same deletion-not-rewrite decision).

- **`code-reviewer`** — required. Standard Go-idiomatic pass over the
  rewritten `pkg/nucleus/*.go`. Pay particular attention to:
  generic `Module[C any].Build()` type-erasure correctness; `Router`
  interface implementation (no method-discovery reflection in
  `Resource`); `func` field reference-identity in `equivalence_test.go`.

- **`security-auditor`** — required. `pkg/nucleus` mounts the same
  defaults `pkg/app` does (CSRF, Casbin default-deny per ADR-004,
  secret redaction per ADR-007). Verify the new builder chain does
  not inadvertently expose `WithoutDefaults()` semantics or skip the
  ADR-004 Casbin mount.

- **`contract-guardian`** — required. Stable surface change. Two
  considerations:
  1. Approve the `pkg/nucleus` baseline reseed
     (`NUCLEUS_UPDATE_CONTRACT_BASELINE=1 go test ./contracts/...`)
     as a *deliberate* contract change, recorded in the iteration log
     and the CHANGELOG `BREAKING (pkg/nucleus rewrite)` entry. This is
     the legitimate-rebaseline path; the freeze test must remain in
     enforcement mode after the reseed.
  2. Confirm no other `pkg/*` baseline changes leak in.

- **`test-runner`** — required. Run lanes in this order:
  1. `go test -race ./pkg/... ./internal/cli ./cmd/nucleus` (race lane,
     no `./examples/...` paths anymore).
  2. `bash scripts/ci/check_contract_freeze.sh` (after pkg/nucleus
     baseline reseed).
  3. `bash scripts/ci/run_compatibility_harness.sh --enforce-threshold`
     (after the harness profiles are scrubbed; ensure the remaining
     profiles still produce a green report).

- **`examples-maintainer`** — **scope inverted for this iteration.**
  Do **not** propose example rewrites. Verify only that:
  1. The `examples/` directory is absent from the working tree.
  2. No `*.go` file outside `docs/iterations/`, `docs/reports/`,
     `docs/audits/` references `examples/*` by import path.
  3. No `*.{md,yml,sh,ps1}` file outside the three historical-archive
     directories above references `examples/*`.
  Report any straggler reference as a blocker and stop the loop.

- **`doc-updater`** — required. Scrub `examples/*` references per the
  Scope §2 list. Update ADR-010 textual revisions per Scope §3. Update
  any `pkg/nucleus` godoc that still describes the legacy chain (the
  inventory in `docs/iterations/2026-05-15-pkg-app-nucleus-inventory.md`
  catalogues the affected symbols). Do **not** edit `docs/iterations/*`,
  `docs/reports/*`, or `docs/audits/*` — those are point-in-time
  archives.

- **`changelog-writer`** — required. Two `BREAKING (...)` entries
  under `### Changed` per Scope §4. Semver impact: **minor → major**
  is the strict-reading answer for pre-`v1.0`, but since the project
  is still pre-`v1.0` the entries land in the next `v0.x.0` minor
  bump (`v0.8.0`) per the ADR-006 / ADR-008 precedent. Note this
  reasoning inline.

- **`dependency-impact`** — not invoked. No `go.mod` changes expected.
  If `go.mod` does change (e.g. removal of an example-only test dep),
  invoke retroactively.

- **`migration-assistant`** — **not invoked.** Pre-`v1.0` clean break.
  No DEP, no MA, no shim. The CHANGELOG `BREAKING (...)` label is the
  whole user-facing migration surface and matches the ADR-006 / ADR-008
  precedent. ADR-010 §263 already commits to this.

- **`performance-bench`** — not invoked. The Phase 1 surface is
  structural; the validator and merge engine (Phase 2) are the first
  hot-path candidates.

- **`governance-checker`** — light-touch pass. Spot-check
  `COMPATIBILITY_SLO.md` cross-reference, freeze-test discipline.
  Full pass deferred to release prep.

- **`session-curator`** — required at session end. Update this file,
  rewrite `HANDOFF.md`, archive into `docs/iterations/YYYY-MM-DD-adr010-phase1-and-examples-purge.md`
  on closure.

## Candidate next steps (re-ordered after this iteration)

1. **`pkg/admin` bootstrap users-table DDL — dialect-aware fix for
   MSSQL/Oracle.** Unchanged from prior queue; still real, still
   blocking `TestSQLMatrix_AutoMigrate_Exploratory` from re-enablement.

2. **ADR-010 Phase 2 — Config loading + merge engine.** Compliance
   items #4, #7, #14, #15, #16, #17. Now the next ADR-010 phase.

3. **ADR-010 Phase 3 — `/_/config` + `nucleus config print --effective`.**
   Compliance items #6, #12, #13.

4. **ADR-010 Phase 4 — Docs-sync + website + new reference
   applications under `examples/`.** Target: v0.9.X. This is where the
   replacement examples get authored.

5. **Cloud Secrets Provider plugin extraction (AWS, then GCP / Azure /
   Vault).** Unchanged from prior queue.

6. **Column-type comparison in `SchemaDrift`.** Unchanged.

7. **SchemaDrift end-to-end usage guide.** Unchanged.

8. **`go mod tidy` unblock.** Unchanged.

9. **`tasks.Manager` struct→interface DEP.** Unchanged.

10. **Audit §7 menores.** Unchanged.

## Files of interest

- `docs/adrs/ADR-010-fluent-api-v2-pkg-nucleus.md` — the Proposed ADR;
  Status flips to Accepted in the Phase 1 PR. Compliance items #1
  (line 275), §263 (line 263), §244 (Compliance §9 #4, line 244)
  receive textual revisions.
- `docs/iterations/2026-05-15-pkg-app-nucleus-inventory.md` — input
  inventory; catalogues the legacy symbols disappearing in Phase 1.
- `pkg/nucleus/nucleus.go` — target of the rewrite.
- `contracts/baseline/api_exported_symbols.txt` — baseline reseeded
  at end of Phase 1.
- `contracts/freeze_test.go:146-164` — `stableAPISymbolBaselineLines`
  package list; already includes `pkg/nucleus`, the baseline lines
  for it are what gets populated.
- `examples/` — entire tree deleted.
- `.github/workflows/{ci,release}.yml`, `scripts/{ci,release,dev}/*`,
  `README.md`, `CHANGELOG.md` — references scrubbed.

## Notes / decisions log

- 2026-05-16 — **Owner decision: delete every `examples/*` tree in
  the same PR as ADR-010 Phase 1 instead of rewriting the two
  `examples/ecommerce_dashboard/backend/*` consumers.** Rationale: the
  existing examples are obsolete relative to the framework's current
  shape; rebuilding them at this point in the cycle adds noise without
  validating the new surface. New reference applications will be
  authored in v0.9.X once Phases 2–4 land. Consistent with pre-`v1.0`,
  single-maintainer, no-external-users posture and the ADR-006 /
  ADR-008 clean-break precedent.

- 2026-05-16 — Side-effect of the examples purge: the compatibility
  harness loses its three integration profiles (`minimal-api`,
  `admin-heavy`, `plugin-heavy`). They are removed from
  `scripts/ci/run_compatibility_harness.sh` for this window. Phase 4
  / v0.9.X restoration is the contract.
