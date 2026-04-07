# Compatibility Harness Report

- Generated at (UTC): 2026-04-07T20:05:06Z
- Branch: `codex/v0.6.0-roadmap`
- Commit: `255e21a`
- Profiles analyzed: 3

| Profile | Status | Duration | Command |
| --- | --- | --- | --- |
| minimal-api | success | 1s | `go test ./examples/mvc_api -run '^TestExampleMVCAPI_Minimal_Smoke$' -count=1 -v` |
| admin-heavy | success | 1s | `go test ./examples/mvc_api -run '^TestExampleMVCAPIAdmin_Smoke$' -count=1 -v` |
| plugin-heavy | success | 1s | `go test ./examples/plugins/... -count=1 -v` |

## Summary

- Passed profiles: 3/3 (100%)
- Threshold: >= 100%
- Decision: READY
