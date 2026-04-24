# Compatibility Harness Report

- Generated at (UTC): 2026-04-24T07:13:49Z
- Branch: `codex/point-4-admin-runtime-impl`
- Commit: `67301fe`
- Profiles analyzed: 3

| Profile | Status | Duration | Command |
| --- | --- | --- | --- |
| minimal-api | success | 3s | `go test ./examples/mvc_api -run '^TestExampleMVCAPI_Minimal_Smoke$' -count=1 -v` |
| admin-heavy | success | 2s | `go test ./examples/mvc_api -run '^TestExampleMVCAPIAdmin_Smoke$' -count=1 -v` |
| plugin-heavy | success | 1s | `go test ./examples/plugins/... -count=1 -v` |

## Summary

- Passed profiles: 3/3 (100%)
- Threshold: >= 100%
- Decision: READY
