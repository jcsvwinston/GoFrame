# Iteration archive — 2026-05-20 contract-freeze scanner constructor-gap fix

> Archived 2026-05-20 as part of the session-end `/handoff`. Shipped
> as PR #77 (`28f75b2`). A short, mechanical governance-tooling
> iteration that ran the full 9-subagent loop.

## Goal

Close the freeze-scanner coverage gap surfaced during the Phase 2d
(#76) review: `contracts/freeze_test.go::exportedSymbolsForPackage`
iterated `go/doc`'s package-level `docPkg.Funcs` but not the per-type
`typ.Funcs`. `go/doc` files a constructor (a function whose result is
the type, e.g. `NewMigrator() *Migrator`) under the returned type's
`Funcs` slice — `docPkg.Funcs` holds only functions `go/doc` could
not associate with any type. As a result every `NewXxx` constructor
across the tracked `pkg/*` set was invisible to the API freeze
baseline: the contract test could not catch a removed or renamed
constructor.

## Scope

### In
- One loop over `typ.Funcs` added inside the type loop of
  `exportedSymbolsForPackage`, emitting `func:Name` (the same encoding
  the top-level function loop uses).
- Reseed `contracts/baseline/api_exported_symbols.txt`.
- A comment at the scanner's `packages` slice documenting the
  deliberately-excluded `pkg/*` packages.

### Out
- No `pkg/*` runtime code; no new framework exports. The constructors
  were always exported and stable — only the scanner's coverage
  changed.

## Acceptance criteria — all met

- [x] `exportedSymbolsForPackage` emits `func:NewXxx` for constructors
      associated with a type via `go/doc`.
- [x] Baseline reseeded **+78 / −0** — purely additive, zero removals.
- [x] `bash scripts/ci/check_contract_freeze.sh` green in enforcement
      mode after the reseed; CI `Contract Freeze` lane green on #77.
- [x] `pkg/db func:NewMigrator` and `pkg/db func:NewModuleMigrator`
      now in the baseline (Phase 2d gap closed).
- [x] Full 9-subagent iteration loop green.

## Status

### Done (2026-05-20, PR #77 → `28f75b2`)

All Scope §In items landed. 78 previously-untracked exported
constructors across the 15 scanned `pkg/*` packages are now frozen
(`db.New`, `NewMigrator`, `NewModuleMigrator`, `router.New`,
`auth.NewJWTManager`, `NewCSRFMiddleware`, `app.New`, `app.Bootstrap`,
`app.LoadConfig`, `authz.New`, `mail.NewSender`, the session-store
constructors, …). No CHANGELOG entry and no semver bump — governance
tooling change with zero user-facing behaviour change (deliberate
decision by changelog-writer + governance-checker).

### Iteration loop

9/9 subagents green:
- **contract-guardian** (key reviewer) PASS — verified +78/−0 additive,
  all entries legitimate exported constructors, no double-counting
  (`go/doc` moves a constructor OUT of `docPkg.Funcs` into `typ.Funcs`,
  so the two loops cover disjoint sets; `dedupeSorted` is a safety net),
  Phase 2d gap closed, CLI/config baselines untouched, no DEP/MA
  warranted.
- **code-reviewer** PASS (NITS — comment-wording suggestions, applied).
- **architect-reviewer** PASS-with-WARN — the WARN is a *separate
  pre-existing gap* (see follow-up below), not this change. Recommended
  merge as-is.
- **test-runner** PASS (contracts + full pkg lane + vet + fmt + build).
- **security-auditor** PASS (test-tooling only; no runtime/security
  surface).
- **examples-maintainer** PASS (no example surface touched).
- **doc-updater** PASS (no doc edit needed — the scanner's `go/doc`
  internals are not a documented contract).
- **changelog-writer** NO-entry (deliberate).
- **governance-checker** PASS (strengthens the SLO's no-removal
  enforcement; no SLO/CI-matrix/release-checklist update needed).

### In progress
- (none)

### Blocked
- (none)

## Follow-up opened by this iteration

**Scanner package-coverage gap** (architect-reviewer WARN). The
scanner's `packages` slice covers 15 `pkg/*` packages but OMITS six:
`pkg/admin`, `pkg/circuit`, `pkg/health`, `pkg/observability`,
`pkg/openapi`, `pkg/outbox`. Their exported surface (including
constructors) has zero removal-protection. A code comment now
documents the deliberate omission at the `packages` slice. Auditing
which to add is candidate #2 in the next `CURRENT_ITERATION.md`
queue — each requires confirming lifecycle posture first (the flat
baseline does not encode lifecycle tags): `pkg/outbox.NewKafkaBridge`
is deliberately unfinished and must NOT be frozen until Kafka delivery
lands; `pkg/openapi` is `experimental`; `pkg/admin` and `pkg/outbox`
are `transitional`.

## Files of interest

- `contracts/freeze_test.go` — the scanner (the `typ.Funcs` loop + the
  `packages`-slice exclusion comment).
- `contracts/baseline/api_exported_symbols.txt` — +78 constructor
  entries.

## Notes / decisions log

- 2026-05-20 — Constructor-gap fix shipped as PR #77 (`28f75b2`).
  Pure governance-tooling change; the freeze test now matches what
  `COMPATIBILITY_SLO.md` already promised ("stable API exported
  symbols no-removal"). Decided against a CHANGELOG entry (no
  user-facing change) and against an ADR (bugfix to an existing
  governance mechanism, covered by the compatibility-by-contract
  ADRs).
