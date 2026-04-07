# Release Checklist

Reference date: 2026-04-05.
Status: Current.

Use this checklist before creating a GoFrame release tag.

## 1. Local Validation

```bash
go test ./...
bash scripts/release/rehearse_rc.sh
```

## 2. Documentation and Changelog

- Ensure `CHANGELOG.md` includes all user-facing changes.
- Ensure README and relevant docs match shipped behavior.

## 3. Version and Tag

- Confirm target version (`v0.x.y` or `v0.x.y-rcN`).
- Create and push tag from a clean `main` commit.

## 4. CI/Release Workflows

Verify:

- CI workflow passes
- release workflow completes
- release asset smoke checks pass

## 5. Compatibility Gates (Mandatory)

Before tagging, attach and review:

- compatibility report (fixture app + stable contract summary)
- exploratory DB stability report (when exploratory lanes are in scope)
- dependency impact report for critical dependencies
- explicit compatibility statement:
  - `no breaking changes`, or
  - `major-only breaking changes with migration plan`

Policy reference:

- `docs/COMPATIBILITY_SLO.md`

## 6. Artifact Review

Check release artifacts include expected OS/arch matrix and checksums.

## 7. Post-Release

- Verify `goframe version` prints the expected release version.
- Update any roadmap/status docs if a phase milestone changed.
