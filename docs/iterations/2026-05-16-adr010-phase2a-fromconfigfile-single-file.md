# Iteration archive — 2026-05-16 ADR-010 Phase 2a: `FromConfigFile` single-file loader

> Archived 2026-05-17 as part of the Phase 1 state-close `/handoff`.
> The Phase 2a iteration completed on 2026-05-16 (PR #73 merged as
> `2b650f3`) but its own state-close did not run in a dedicated session.
> The convention established by #61 / #64 / #68 (state-close PRs follow
> the feature PR by one commit) is preserved by capturing the iteration
> record here.

## Goal

Replace the Phase 1 `ErrConfigLoaderNotImplemented` stub in
`pkg/nucleus.AppBuilder.FromConfigFile` with the first slice of the
real five-layer loader described in ADR-010 §2: a **single-file YAML
loader** that enforces the §17 size cap, performs schema-level
strict-unknown-keys validation against `app.ContractConfigKeyPatterns()`,
and surfaces structured errors with did-you-mean hints. The
multi-file merge engine and the TOML / JSON parsers are sliced into
Phase 2b; the production-strict guard is Phase 2c; module migration
namespacing is Phase 2d.

## Scope

### In

- `pkg/nucleus/config.go` (new, 315 lines) — single-file YAML loader.
  - **1 MiB per-file size cap** (`MaxConfigFileBytes`) via
    `io.LimitReader(file, cap+1)` so overshoot is detected by the
    read, not by a potentially-lying `os.Stat` (procfs / FUSE caveats).
    Returns `ErrConfigFileTooLarge` before the YAML parser runs —
    eliminates anchor-expansion / deep-nesting DoS classes against
    `gopkg.in/yaml.v3`.
  - **Extension-based parser inference**: `.yaml` / `.yml` use the
    existing koanf YAML provider; `.toml` / `.json` return
    `ErrUnsupportedConfigFormat` with a Phase 2b pointer.
  - **Strict-unknown-fields schema validation** against
    `app.ContractConfigKeyPatterns()`. Two koanf instances: one to
    enumerate the file's keys for the strict check, the second for
    the layered load (defaults < file). Unknown keys surface as
    `ErrUnknownConfigKeys` listing every offending key with
    did-you-mean hints (Levenshtein distance ≤3 on the final
    segment).
  - **Wildcard pattern matching**: `keyMatchesAny` recognises
    map-typed schema slots like `databases.*.url` and
    `jwt_keys.*.kid`.

- `pkg/nucleus/config_test.go` (new, 274 lines) — 13 cases covering
  happy path, defaults preservation, unsupported extensions,
  Phase 2b TOML/JSON sentinel, cap-boundary handling, unknown-key
  rejection, did-you-mean hints, missing-file vs content-error
  distinction, malformed YAML, empty path, end-to-end builder
  integration, Modules preserved across `FromConfigFile`, Levenshtein
  basics, wildcard matcher.

- `pkg/nucleus/nucleus.go` updates (+37 / −38 lines):
  - `AppBuilder.FromConfigFile(path)` now invokes the real loader.
  - Multi-path `FromConfigFile(a, b)` fails fast with a Phase 2b
    reference (`ErrUnsupportedConfigFormat` family — merge engine is
    the next sub-PR).
  - Modules / Middleware / Services / Lifecycle registered BEFORE
    the `FromConfigFile` call are preserved; only the embedded
    `app.Config` slot is replaced. Regression test included.
  - `ErrConfigLoaderNotImplemented` is removed entirely (Phase 1
    stub retired). Pre-`v1.0` clean break per the ADR-006 /
    ADR-008 precedent.

- `pkg/nucleus/equivalence_test.go` (+11 / −13 lines) — the three-surface
  equivalence test updated for the new `FromConfigFile` semantics.

- Freeze-baseline rebase: net delta **+4 / −1** in
  `contracts/baseline/api_exported_symbols.txt`. New entries:
  `const MaxConfigFileBytes`, `var ErrConfigFileTooLarge`,
  `var ErrUnsupportedConfigFormat`, `var ErrUnknownConfigKeys`.
  Removed: `var ErrConfigLoaderNotImplemented` (Phase 1 stub).

- `CHANGELOG.md` `Unreleased` entry under `### Added` describing the
  new loader and the four new exported names.

- `go.mod` / `go.sum`: new dependency
  `github.com/knadh/koanf/providers/rawbytes v1.0.0` (sibling of the
  YAML provider already in tree; zero-cgo; feeds the read bytes into
  koanf without going through the filesystem twice).

### Out (deferred to later sub-iterations)

- **Phase 2b** — multi-file merge engine with `_append` / `_remove`
  suffix operators (ADR-010 §3); TOML / JSON parsers; non-nullable
  security keys (`cors.origins`, `auth.providers`,
  `authz.policy_path`, `session.secret`) per ADR-010 §14.
- **Phase 2c** — `WithUnknownFields("warn")` opt-out;
  `NUCLEUS_ENV=production` strict override; the startup `WARN` line
  per ADR-010 §15.
- **Phase 2d** — module migration namespacing in `pkg/db/migrate.go`
  per ADR-010 §16.

## Acceptance criteria (all met)

- [x] `pkg/nucleus/config.go` implements the single-file YAML loader.
- [x] 1 MiB size cap enforced before YAML parse via `io.LimitReader`.
- [x] Strict unknown-key validation against
      `app.ContractConfigKeyPatterns()` with did-you-mean hints.
- [x] `.toml` / `.json` paths return `ErrUnsupportedConfigFormat`
      with a Phase 2b pointer.
- [x] `AppBuilder.FromConfigFile(path)` end-to-end integration test
      passes; Modules mounted before the call are preserved.
- [x] Multi-path `FromConfigFile(a, b)` fails fast with a Phase 2b
      reference (merge engine deferred).
- [x] `ErrConfigLoaderNotImplemented` Phase 1 stub removed; freeze
      baseline rebased (+4 / −1 net).
- [x] `pkg/nucleus/config_test.go` covers 13 cases (see Scope §In).
- [x] CI green: 11/11 checks SUCCESS on PR #73.

## Status

### Done (2026-05-16, PR #73 → `2b650f3`)

All Scope §In items landed. Phase 2a closed.

### In progress

(none)

### Blocked

(none)

## Files of interest

- `pkg/nucleus/config.go` — the loader.
- `pkg/nucleus/config_test.go` — 13 test cases.
- `pkg/nucleus/nucleus.go` — builder integration.
- `contracts/baseline/api_exported_symbols.txt` — rebased (+4 / −1).
- `CHANGELOG.md` — `Unreleased` entry under `### Added`.
- `docs/adrs/ADR-010-fluent-api-v2-pkg-nucleus.md` — §17 (file-size
  cap) and §2 (validation layers 1-2) are now satisfied; §3 (merge
  engine), §14-§16 remain for Phase 2b / 2c / 2d.

## Notes / decisions log

- 2026-05-16 — Owner sliced ADR-010 §2 (originally "Phase 2 — config
  loading + merge engine") into four sub-iterations 2a / 2b / 2c / 2d
  so each lands as its own reviewable PR. 2a is the file loader +
  size cap + schema validation; 2b is the merge engine + extra
  parsers; 2c is the production-strict guard; 2d is module
  migration namespacing.

- 2026-05-16 — `io.LimitReader(file, cap+1)` chosen over `os.Stat()`
  for the size cap because procfs and FUSE filesystems can report
  arbitrary sizes via `Stat` while the actual byte stream tells the
  truth. The `+1` byte allows the loader to distinguish "exactly at
  the cap" (accepted) from "over the cap" (rejected with
  `ErrConfigFileTooLarge`).

- 2026-05-16 — Two koanf instances used during load: one for the
  strict-unknown-keys enumeration of the file, one for the layered
  load (defaults < file). This avoids polluting the validation
  surface with defaulted keys that wouldn't otherwise appear in the
  file.

- 2026-05-16 — Did-you-mean hint uses Levenshtein distance ≤3 on the
  final segment of the dotted key. Final-segment-only is cheaper to
  compute and produces useful hints for the typical typo class
  (e.g. `databasees.default.url` → `databases.default.url`).

- 2026-05-17 — Phase 2a state-close archived as part of the Phase 1
  state-close `/handoff` (combined to avoid a separate stub PR for a
  single-iteration archive). Phase 1 and Phase 2a both completed on
  2026-05-16; archives carry that completion date in their filename.
