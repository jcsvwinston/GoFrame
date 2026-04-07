# Versioning Strategy

Reference date: 2026-04-05.
Status: Current.

GoFrame follows Semantic Versioning while in pre-1.0 mode.

## Current Policy

Version format:

- `v0.x.y`

Interpretation in pre-1.0:

- `x` (minor): may include significant feature additions and limited breaking changes
- `y` (patch): bug fixes, hardening, and non-breaking improvements

Pre-1.0 note:

- While breaking changes are still technically possible before `v1.0`, they should be treated as exceptions and require explicit migration notes.
- Strategic direction for the `v1.x` era is defined in `docs/LONG_TERM_COMPATIBILITY_ROADMAP.md`.

## v1.x Compatibility Commitment (Target)

From `v1.0` onward:

- no breaking changes in `v1.x` for stable public contracts
- deprecations must provide migration path and tooling before any major-version removal

## Release Types

1. Release candidates
- Format: `v0.x.y-rcN`
- Used to validate release packaging and workflows before stable promotion

2. Stable pre-1.0
- Format: `v0.x.y`
- Promoted after CI, rehearsal, and artifact checks pass

## Source of Truth

- Git tags are the version source of truth.
- Binary version output is injected at build time.

## Required Checks Before Tagging

```bash
go test ./...
bash scripts/release/rehearse_rc.sh
```

## Changelog Discipline

Every user-facing change should be reflected in `CHANGELOG.md` under `Unreleased` before release.
