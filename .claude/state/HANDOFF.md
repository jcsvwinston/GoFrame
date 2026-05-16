# Handoff — last session closing note

> Owned by `session-curator`. Overwritten at the end of every session
> by `/handoff`. Read first by `/resume` at the start of the next one.

ITERATION:    ADR-010 Phase 1 (Fluent API v2 Foundation) + wholesale removal of `examples/*`. REGISTERED, not yet started. See `.claude/state/CURRENT_ITERATION.md` for the full scope, acceptance criteria, and per-subagent guidance.
BRANCH:       claude/thirsty-matsumoto-8c6d9a (worktree). `main` is at c59d775 (PR #69 — ADR-010 Phase 0 revised draft, merged 2026-05-16).
LAST COMMIT:  c59d775 docs(adr): ADR-010 Phase 0 — revised draft after subagent review (#69)
STATUS:       Iteration registered today. Two state files written in this worktree (CURRENT_ITERATION.md, HANDOFF.md); no code or docs touched yet.
NEXT STEP:    Run `/iterate` to begin Phase 1. The iteration loop owns its work; the per-subagent scope guidance in `CURRENT_ITERATION.md` §"Subagent guidance for this iteration" pre-resolves the two scope inversions (`examples-maintainer` verifies deletion, `migration-assistant` is not invoked) so the loop does not have to rediscover them.
BLOCKERS:     none.
FILES OF INTEREST:
  - .claude/state/CURRENT_ITERATION.md — the authoritative scope and acceptance criteria for this iteration.
  - docs/adrs/ADR-010-fluent-api-v2-pkg-nucleus.md — Status currently Proposed; flips to Accepted in this PR. §263, §244 (Compliance §9 #4), §297, Compliance #1 receive textual revisions to reflect the deletion-not-rewrite decision.
  - docs/iterations/2026-05-15-pkg-app-nucleus-inventory.md — input inventory.
  - pkg/nucleus/nucleus.go — target of the rewrite.
  - contracts/baseline/api_exported_symbols.txt — baseline reseeded at end of Phase 1.
  - examples/ — entire tree deleted; references scrubbed across .github/workflows, scripts/, docs/ (excluding historical archives under docs/iterations, docs/reports, docs/audits), README.md.

NOTES:
  - **Owner decision (2026-05-16):** every `examples/*` tree is deleted in this PR rather than rewriting the two `examples/ecommerce_dashboard/backend/*` consumers ADR-010 §263 originally specified. Existing examples are obsolete; new reference applications will be authored in v0.9.X (Phase 4 / docs-sync iteration). Consistent with pre-`v1.0`, single-maintainer, no-external-users posture and the ADR-006 / ADR-008 clean-break precedent.
  - Side-effect: `scripts/ci/run_compatibility_harness.sh` loses its `minimal-api`, `admin-heavy`, `plugin-heavy` profiles for this window. They get restored together with the new examples in v0.9.X.
  - The state-file write happened in worktree `claude/thirsty-matsumoto-8c6d9a`; commit + PR + merge of these two files to `main` is the first concrete action of the iteration (they need to land before `/iterate` will read them on a fresh clone).

OPEN HOUSEKEEPING (none blocking, carried from prior sessions):
  - `go mod tidy` cannot run cleanly (pre-existing admin/proto replace-directive issue) — AWS SDK modules show as `// indirect`. Will become moot once the Cloud Secrets plugin extraction lands.
  - `panic(` count in non-test code reportedly 4→0 since b1e497e — still unconfirmed; worth a quick verification pass in a quiet session.

Updated: 2026-05-16
