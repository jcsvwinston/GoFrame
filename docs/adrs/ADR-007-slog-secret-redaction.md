# ADR-007: Secret Redaction in the Structured Logger

**Status:** Accepted
**Date:** 2026-05-14
**Superseded:** No

## Context

The 2026-05-14 post-sprint readiness audit (§7 item 6) flagged that `pkg/observe`'s logger does no secret redaction. `NewLogger` built a `slog.Handler` with `slog.HandlerOptions{Level: lvl}` and nothing else — no `ReplaceAttr`. Any code that logged a secret-bearing attribute emitted it verbatim:

```go
logger.Info("auth failed", "authorization", r.Header.Get("Authorization"))
// → {"level":"INFO","msg":"auth failed","authorization":"Bearer eyJ..."}
```

The framework already redacts on the *event-bus* side (`pkg/observability/hooks/{http,sql}.go` strip query strings and SQL parameters), but that is a separate pipeline. Direct `slog` calls — the overwhelmingly common logging path, used throughout `pkg/*`, `internal/cli`, and consumer applications — had no protection. A bearer token, a session cookie, a DB password logged in a debug line lands in plaintext in whatever log sink the deployment ships to.

This violates SPEC.md §2 principle 4 (security-by-default). It is the sibling defect to the CSRF gaps closed in ADR-006: the same audit, the same "the framework hands you a footgun by default" shape.

`NewLogger(level, format string) *slog.Logger` is on the **stable** contract surface (`pkg/observe` is in `contracts/baseline/api_exported_symbols.txt`). Its signature cannot change. Its *behaviour* can — and that is the ADR-worthy decision.

## Decision

### 1. Redaction is ON by default

`NewLogger` now installs a `ReplaceAttr` hook that redacts the value of any attribute whose key is in a built-in denylist. The value is replaced with `RedactionPlaceholder` (`"[REDACTED]"`); the key is kept so the shape of the log line is unchanged and the redaction is visible and greppable.

Security-by-default means the *default* constructor is the *safe* one. An operator does not opt in to not leaking secrets — they would have to opt **out**, explicitly, in code.

### 2. The denylist is curated and exact-match

`defaultRedactedKeys` is a curated set — `authorization`, `cookie`, `set-cookie`, `password`, `token`, `secret`, `api_key`, `access_token`, `private_key`, … (the full set is returned by the exported `DefaultRedactedKeys()` so operators can audit it). Matching is **case-insensitive exact match** on the attribute key.

Substring / suffix matching (`*_token`, `*_key`, `*_secret`) was considered and rejected. It catches more secrets but also redacts benign fields — `page_token`, `cache_key`, `partition_key`, `sort_key` — silently hiding debugging information. A predictable exact-match list that operators extend for their own fields beats a clever pattern that surprises them. The ADR records this so the choice is not silently re-litigated.

### 3. Customisation is code-level, not config-level

A new additive constructor `NewLoggerWithRedaction(level, format string, cfg RedactionConfig)` gives explicit control:

- `RedactionConfig.ExtraKeys` — additional app-specific keys to redact (`ssn`, `card_number`, …).
- `RedactionConfig.Placeholder` — override the redacted-value string.
- `RedactionConfig.Disabled` — turn redaction off entirely.

There is deliberately **no config-file key** to disable redaction. Disabling a security default requires touching code (`NewLoggerWithRedaction` with `Disabled: true`), which surfaces in code review — the same discipline ADR-004 applies to `app.WithOpenAuthz()`. A config flag would let a deployment silently turn the protection off; a code change cannot hide.

`ExtraKeys` *is* reachable from configuration: `App.New` reads the `log_redact_extra_keys` config key (registered in `CONFIG_KEY_REGISTRY.md`, lifecycle `transitional`) and threads it into `NewLoggerWithRedaction`. Extending the denylist is additive and safe, so it gets a config surface; removing protection does not.

### 4. `NewLogger` delegates; the writer is injectable internally

`NewLogger` is now `NewLoggerWithRedaction(level, format, RedactionConfig{})`. Both delegate to an unexported `newLogger(w io.Writer, …)` so tests can capture and assert on the actual rendered output. The exported constructors always pass `os.Stdout`; the `io.Writer` seam is internal and not part of the contract.

### Alternatives considered

- **Opt-in redaction.** Rejected — defeats security-by-default; the audit finding is precisely that the default is unsafe.
- **Change `NewLogger`'s signature to take options.** Rejected — breaks the frozen `stable` signature for no benefit over the additive `NewLoggerWithRedaction`.
- **A mutable package-level denylist global.** Rejected — SPEC.md §2 principle 2 ("no hidden globals"); a global is not safe for concurrent configuration and makes the redaction behaviour non-local and hard to reason about.
- **Redact by value pattern-matching** (detect things that *look* like JWTs / keys in any field). Rejected for v0.7.x — high false-positive/false-negative rate, expensive on the log hot path. Key-based redaction is cheap and predictable; value-scanning can be revisited later if needed.

## Consequences

### Positive

- A secret logged through a sensibly-named attribute (`password`, `authorization`, `token`, …) no longer reaches the log sink in plaintext — by default, with zero operator action.
- The redacted line keeps its key and structure, so log-shape-dependent tooling (dashboards, parsers) is unaffected; the value is just `[REDACTED]`.
- Operators can audit exactly what is redacted (`DefaultRedactedKeys()`), extend it (`ExtraKeys` / `log_redact_extra_keys`), or — with an explicit, review-visible code change — disable it.

### Negative

- **Behaviour change on a stable surface.** A deployment that *intentionally* logged a field named, say, `token` (perhaps an opaque non-secret correlation token) now sees `[REDACTED]` there. This is the correct trade-off — a field named `token` is a secret far more often than not — but it is a visible change. Documented in `CHANGELOG.md` under `Changed`; the escape hatch (`NewLoggerWithRedaction` with `ExtraKeys` excluded, or a renamed attribute) is documented in `OBSERVABILITY_BASELINE.md`.
- The exact-match list will miss secrets logged under non-obvious keys (`auth_header`, `bearer`, a typo). Redaction is a safety net, not a substitute for not logging secrets in the first place. The guide says so explicitly.
- A tiny per-attribute cost on the log path (one lower-case + one map lookup per attribute). Negligible, and zero when redaction is disabled (`ReplaceAttr` is left unset).

### Neutral

- The event-bus redaction in `pkg/observability/hooks` is unchanged and untouched — it is a different pipeline with different semantics (query-string / SQL-parameter stripping). Unifying the two is out of scope.

## Compliance

After this ADR is accepted:

1. `pkg/observe.NewLogger` redacts by default via a `ReplaceAttr` hook.
2. `RedactionConfig`, `NewLoggerWithRedaction`, `DefaultRedactedKeys`, and `RedactionPlaceholder` are exported and added to `contracts/baseline/api_exported_symbols.txt`.
3. There is no config key that disables redaction. `log_redact_extra_keys` (extend-only) is registered in `docs/reference/CONFIG_KEY_REGISTRY.md`, lifecycle `transitional`, and wired through `App.New`.
4. `docs/guides/OBSERVABILITY_BASELINE.md` documents the default denylist, the `ExtraKeys` / `Disabled` / `Placeholder` knobs, the `log_redact_extra_keys` config key, and the "redaction is a safety net, do not log secrets" caveat.
5. `CHANGELOG.md` records the constant — redaction-by-default — under `Security`, and the stable-surface behaviour change under `Changed`.

## Related

- [`pkg/observe/redact.go`](../../pkg/observe/redact.go), [`pkg/observe/logger.go`](../../pkg/observe/logger.go).
- `docs/audits/2026-05-14-post-sprint-readiness.md` §7 item 6 — the audit finding.
- ADR-001: stdlib-first — `slog.HandlerOptions.ReplaceAttr` is the stdlib mechanism; no dependency added.
- ADR-004: Casbin default-deny — precedent for "security default, code-level opt-out, no config switch to disable".
- ADR-006: CSRF hardening — the sibling security fix from the same audit.
- `SPEC.md` §2 — principle 2 (no hidden globals) and principle 4 (security-by-default).
