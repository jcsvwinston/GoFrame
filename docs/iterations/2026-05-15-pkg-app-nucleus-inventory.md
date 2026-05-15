# Inventory — 2026-05-15 — `pkg/app` + `pkg/nucleus` for Fluent API v2

**Purpose:** Read-only Phase 1 inventory for an upcoming Fluent API v2
ADR on `pkg/nucleus`. Produced by `architect-reviewer` and
`contract-guardian` dispatched in parallel. **No code changes** were
made; the decision between in-place rewrite and `pkg/nucleus/v2`
coexistence is **not** recorded here and belongs to the upcoming ADR
(or an ADR-005 addendum).

**Scope of inputs:**

- `pkg/app/*.go` (19 source files) — full read.
- `pkg/nucleus/{context,nucleus,routes}.go` and their `_test.go`
  counterparts — full read.
- `SPEC.md §3.1 — Application Container (pkg/app)` — full read.
- `contracts/baseline/api_exported_symbols.txt` and
  `contracts/freeze_test.go` — full read.
- `examples/**` — grep for `jcsvwinston/nucleus/pkg/nucleus`.
- `docs/governance/COMPATIBILITY_SLO.md`, ADR-006, ADR-008 — read for
  pre-`v1.0` BREAKING precedent.

**Key facts surfaced during the pass:**

1. `pkg/nucleus` has **zero** entries in
   `contracts/baseline/api_exported_symbols.txt` (`grep` returns
   nothing).
2. `pkg/nucleus` is **not enumerated** in the hardcoded package list at
   [contracts/freeze_test.go:146-164](contracts/freeze_test.go:146).
   The freeze-test scanner does not currently observe this package —
   so it cannot detect removals or renames there. That is itself a
   finding; it shapes both rewrite paths.
3. Only **two consumer files** import `pkg/nucleus`:
   `examples/ecommerce_dashboard/backend/main.go` and
   `examples/ecommerce_dashboard/backend/handlers/handlers.go`. No
   usage in `cmd/`, `internal/`, or any other `pkg/*`.
4. `pkg/nucleus` carries multiple latent defects that affect runtime
   behaviour and must be addressed by either rewrite strategy:
   the dead `b.router` field, the silently-dropped `Cors()` middleware,
   the range-copy bug in `Use()`, the panic-on-error `Load()`, the
   no-op `AutoMigrate()` method, and a plain `slog.TextHandler` logger
   that bypasses ADR-007 secret redaction.

---

## 1. `pkg/app` + `pkg/nucleus` layering map

### 1.1 `pkg/app` (the application container — SPEC.md §3.1)

`pkg/app` is the highest-level package in the dependency tree. It
imports many `pkg/*` siblings and is imported by `pkg/nucleus`, by
`cmd/nucleus`, by tests, and by examples — but does **not** import
any `internal/*` package and is **not** imported by any other `pkg/*`.

Per-file roles:

- **[pkg/app/app.go](pkg/app/app.go)** — `App` struct + lifecycle. The
  contract surface SPEC.md §3.1 describes lives here:
  - `App.Config`, `App.Logger`, `App.Router`, `App.DB`, `App.DBs`,
    `App.Mailer`, `App.Session`, `App.JWT`, `App.Models`, `App.Admin`,
    `App.Authorizer`, `App.Storage`, `App.Outbox`, `App.Templates`,
    `App.Observability`, `App.SessionRecorder`.
  - `New(cfg *Config, opts ...Option) (*App, error)` — constructor;
    wires every subsystem via `attachDefaultSubsystems` unless
    `WithoutDefaults()` is set.
  - `Run(ctx context.Context) error` — starts the HTTP server, attaches
    `SIGINT/SIGTERM` via `signal.Notify`
    ([pkg/app/app.go:971](pkg/app/app.go:971)), blocks until cancel
    signal.
  - `Shutdown(ctx context.Context) error` — drains shutdown hooks in
    **reverse** registration order
    ([pkg/app/app.go:1016-1019](pkg/app/app.go:1016)).
  - `OnShutdown(fn func(context.Context) error)` — registers a
    shutdown hook
    ([pkg/app/app.go:920](pkg/app/app.go:920)).
  - `AutoMigrate`, `RegisterModel`, `MountAdmin`, `MountOpenAPI`,
    `Database`, `DatabaseForRequest`, `DefaultDB`,
    `DefaultDatabaseAlias`.
- **[pkg/app/config.go](pkg/app/config.go)** — `Config` struct with
  `koanf`-tagged fields, `LoadConfig(path ...string)`, `DefaultConfig`,
  internal `defaults()` and normalisation helpers.
- **[pkg/app/extensions.go](pkg/app/extensions.go)** — `Extension`
  interface (`Name()`, `Attach(a *App) error`, `Shutdown(ctx)`), the
  `Option` func type, `WithExtensions`, `WithoutDefaults`,
  `WithOpenAuthz`.
- **[pkg/app/bootstrap.go](pkg/app/bootstrap.go)** — `Bootstrap`
  (loads config then `New`) and `QuickStart` (calls `log.Fatalf` on
  error — a long-standing design smell, not a blocker for v2).
- **[pkg/app/requestscope.go](pkg/app/requestscope.go)** —
  `RequestScope`, `RequestScopeFromContext`, `SiteFromContext`,
  `TenantFromContext`, `DatabaseAliasFromContext`.
- **[pkg/app/errors.go](pkg/app/errors.go)** — sentinel errors +
  `OpError` wrapper, stdlib-only imports.
- **[pkg/app/healthz.go](pkg/app/healthz.go)** — `/healthz` handler;
  always registered, regardless of `WithoutDefaults`
  ([pkg/app/app.go:415](pkg/app/app.go:415)).
- **[pkg/app/authz_default.go](pkg/app/authz_default.go)** — default-deny
  RBAC middleware (ADR-004).
- **[pkg/app/jwt_setup.go](pkg/app/jwt_setup.go)** — `buildJWTManager`;
  nil-returns when no signing material configured
  ([pkg/app/jwt_setup.go:58-65](pkg/app/jwt_setup.go:58)).
- **[pkg/app/config_contract.go](pkg/app/config_contract.go)** —
  `ContractConfigKeyPatterns() []string` for the contract-freeze test.

Outward imports from `pkg/app`: `pkg/admin`, `pkg/auth`,
`pkg/auth/secrets`, `pkg/authz`, `pkg/db`, `pkg/errors`, `pkg/health`,
`pkg/mail`, `pkg/model`, `pkg/observe`, `pkg/observability`,
`pkg/observability/hooks`, `pkg/openapi`, `pkg/outbox`, `pkg/router`,
`pkg/storage`. All one-way (no `pkg/*` imports `pkg/app`).

**Hidden coupling in `pkg/app`:**

- `init()` in `pkg/app/integration_sprint_test.go:41` registers a mail
  driver via `mail.RegisterProvider` — test-only, not shipped.
- `pkg/app/app.go:377` calls `model.SetDefaultSQLObserver(...)` inside
  `New()`, mutating a process-wide singleton in `pkg/model`. First
  `App.New` call wins. Today `App` is conceptually a singleton so this
  is not observable; a v2 design that allows two `App` instances to
  coexist (e.g. in-test parallel construction) would surface this.

### 1.2 `pkg/nucleus` (the fluent facade)

Three source files, all in one package; no subpackages.

- **[pkg/nucleus/nucleus.go](pkg/nucleus/nucleus.go)** — `AppBuilder`
  struct, fluent entry (`New`, `Load`), config-mutation chain methods
  (`Port`, `Host`, `SQLite`, `Postgres`, `MySQL`, `WithAdmin`,
  `Templates`, `Static`, `Cors`, `Provide`, `Model`, `AutoMigrate`,
  `SPA`), escape hatches (`WithConfig`, `WithConfigAny`, `Config`,
  `Logger`), terminal `Run`.
- **[pkg/nucleus/context.go](pkg/nucleus/context.go)** — `Context`
  struct embedding `*routerpkg.Context`, plus bind/respond/session
  helpers.
- **[pkg/nucleus/routes.go](pkg/nucleus/routes.go)** — `Handler` and
  `Middleware` type aliases, top-level route registration
  (`Get`/`Post`/`Put`/`Delete`/`Group`/`Resource`/`Use`), `RouterGroup`
  and its method set.

**What `pkg/nucleus` adds on top of `pkg/app`:**

- A builder pattern (fluent chain) over `app.Config`.
- A simplified `Handler` signature `func(*Context) error` that wraps the
  stdlib `http.Handler` interface.
- A `Context` type that adds `BindJSON`/`BindXML`/`BindForm`/`JSON`/
  `XML`/`HTML`/`String`/etc. on top of `routerpkg.Context`.
- An `SPAConfig` and `(*AppBuilder).SPA` for client-side-routing
  fallback (no equivalent in `pkg/app`).

**Call graph at runtime:** `nucleus.New()` → builder mutation chain →
`(*AppBuilder).Run()` → `app.New(&b.config)` → `(*App).Run(context.Background())`
([pkg/nucleus/nucleus.go:205](pkg/nucleus/nucleus.go:205)).

**Architectural defects observed in the current `pkg/nucleus`** (each
must be addressed by either rewrite strategy):

- **Dead `b.router` field.** `New()` constructs a
  `routerpkg.New(logger)` at
  [pkg/nucleus/nucleus.go:51](pkg/nucleus/nucleus.go:51) and stores it
  in `b.router`. `Run()` calls `app.New(&b.config)` which constructs a
  fresh router internally — `b.router` is then discarded silently.
- **Broken `Cors()` middleware.** `(*AppBuilder).Cors` at
  [pkg/nucleus/nucleus.go:155](pkg/nucleus/nucleus.go:155) adds the
  middleware to `b.router` — the discarded router. The CORS
  configuration entered via the fluent chain is **silently dropped at
  runtime**.
- **Broken `Use()` middleware injection.**
  [pkg/nucleus/routes.go:203-210](pkg/nucleus/routes.go:203) iterates
  `b.routes` by value (Go `range` copies each element); mutations to
  `route.middlewares` inside the loop never persist back to the stored
  slice. `Use()` is effectively a no-op.
- **Panic-on-error `Load()`.**
  [pkg/nucleus/nucleus.go:69](pkg/nucleus/nucleus.go:69) panics on a
  config load failure. Violates SPEC §2 Principle 2 (explicit
  lifecycle).
- **No-op `AutoMigrate()`.**
  [pkg/nucleus/nucleus.go:197-199](pkg/nucleus/nucleus.go:197) does
  nothing — the real auto-migration is triggered by a `len(b.models)>0`
  check inside `Run()`. The method name is misleading.
- **Plain `slog.NewTextHandler` logger bypasses ADR-007 redaction.**
  [pkg/nucleus/nucleus.go:45](pkg/nucleus/nucleus.go:45) creates the
  logger directly. `app.New` later builds the redaction-aware logger,
  but anything logged through `b.logger` between `New()` and `Run()`
  (e.g. config-load errors) goes through the plain handler. A
  secret-bearing config field logged via `b.logger` would be emitted
  in plaintext.
- **Missing `WithoutDefaults` / `WithExtensions` seam.** The fluent
  API has no way to opt out of the default-deny RBAC, mail wiring,
  storage wiring, or admin panel — operators who want the lightweight
  `app.New(cfg, app.WithoutDefaults())` core-only path cannot reach
  it through `AppBuilder`.

---

## 2. Cycle and coupling audit

**Import direction.** Strictly `pkg/nucleus → pkg/app → (many
pkg/*)`. `pkg/app` does not import `pkg/nucleus` (confirmed by grep
returning nothing).

**Import cycles.** None today. The direction cannot invert without
creating one.

**Test cycles.**

- `pkg/app/*_test.go`: no imports of `pkg/nucleus`. The lone `init()`
  in `pkg/app/integration_sprint_test.go:41` registers a test mail
  driver and is self-contained.
- `pkg/nucleus/*_test.go`: imports `pkg/router` (appropriate — `Context`
  embeds `*routerpkg.Context`) but not `pkg/app`. Tests are in
  `package nucleus` so they can read unexported fields (`builder.config.Port`
  etc.).

**Hidden coupling.** Two material items:

1. `(*AppBuilder).Run()` calls `app.New(&b.config)` with **no
   options** — the entire `WithoutDefaults`/`WithExtensions` seam is
   invisible from the fluent API. Operators on `pkg/nucleus` get the
   full default stack unconditionally.
2. `app.New` mutates the process-wide `model.SetDefaultSQLObserver`
   ([pkg/app/app.go:377](pkg/app/app.go:377)). Today `App` is
   functionally a singleton, so this is invisible; a v2 design that
   permits two `App` instances per process (e.g. for parallel tests)
   would race here.

---

## 3. SPEC.md §3.1 contract surface that v2 must preserve

SPEC.md §3.1 — "Application Container (`pkg/app`)" — defines the
guarantees any v2 fluent layer must keep, because both rewrite paths
delegate to `app.New`. Each row maps a SPEC.md guarantee to the
`pkg/app` symbol that implements it today.

| SPEC §3.1 guarantee | Implementing `pkg/app` symbol |
|---|---|
| Config loaded + normalised + multi-tenant-isolated at construction time | `app.New` → `mergeDefaults` + `validateMultiTenantIsolation` ([pkg/app/app.go:211-212](pkg/app/app.go:211)) |
| `log/slog` logger with ADR-007 secret redaction enabled by default | `observe.NewLoggerWithRedaction` ([pkg/app/app.go:225](pkg/app/app.go:225)) |
| SQL database map opened per alias; default alias resolved | `openDatabases` ([pkg/app/app.go:1237](pkg/app/app.go:1237)) |
| Session manager initialised (memory/sql/redis); session middleware installed | `buildSessionManager` + `r.Use(sessionManager.Middleware())` ([pkg/app/app.go:312-334](pkg/app/app.go:312)) |
| Request scope resolver middleware installed | `r.Use(scopeResolver.Middleware())` ([pkg/app/app.go:333](pkg/app/app.go:333)) |
| `Observability` bus always non-nil | `observability.NewBus` ([pkg/app/app.go:363](pkg/app/app.go:363)) |
| `JWT` nil when no signing material configured | `buildJWTManager` nil-return guard ([pkg/app/jwt_setup.go:58-65](pkg/app/jwt_setup.go:58)) |
| Extension lifecycle: `Attach` in registration order; `Shutdown` in reverse | `app.go:488-495` attach loop + `app.go:1016-1019` reverse shutdown loop |
| `WithoutDefaults()` opt-out from admin/storage/mail/authz | `extensions.go:64-68` + `app.go:472-477` |
| `Run`/`Shutdown` with hook registration via `OnShutdown` | `app.Run` ([pkg/app/app.go:930](pkg/app/app.go:930)), `app.Shutdown` ([pkg/app/app.go:992](pkg/app/app.go:992)), `app.OnShutdown` ([pkg/app/app.go:920](pkg/app/app.go:920)) |
| Database close registered as shutdown hook | `a.OnShutdown(closeDatabases)` ([pkg/app/app.go:460](pkg/app/app.go:460)) |
| Telemetry shutdown registered | `a.OnShutdown(telemetryShutdown)` ([pkg/app/app.go:463](pkg/app/app.go:463)) |
| Admin bootstrap user created at startup, not lazily | `admin.EnsureBootstrapAdminUser` ([pkg/app/app.go:268](pkg/app/app.go:268)) |
| `/healthz` always registered (even with `WithoutDefaults`) | `a.Router.Get("/healthz", a.handleHealthz)` ([pkg/app/app.go:415](pkg/app/app.go:415)) |
| SIGINT/SIGTERM trigger graceful shutdown | `signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)` ([pkg/app/app.go:971](pkg/app/app.go:971)) |

Any v2 design that builds its own lifecycle — bypassing `app.New` —
must replicate every row. Designs that keep `app.New` as the
under-the-hood implementation inherit them automatically; the v2
question becomes purely one of ergonomic surface, not lifecycle
correctness.

---

## 4. Contract surface, consumers, and freeze-test impact

### 4.1 Exported-symbol inventory and classification

Every exported symbol in `pkg/nucleus` and its classification by the
contract-guardian pass. Classifications:

- **legacy chain method** — exists only so the chain reads well;
  underlying functionality is a one-line `pkg/app` setter. Candidate
  for removal in v2.
- **still useful** — the fluent form is the only ergonomic path;
  v2 must keep it (possibly renamed).
- **unsupported** — declared but unused, dead code, or a no-op
  wrapper.

| Symbol | Kind | Where | Classification |
|---|---|---|---|
| `AppBuilder` | struct | `nucleus.go:21` | still useful (chain accumulator) |
| `SPAConfig` | struct | `nucleus.go:38` | still useful (no `pkg/app` equivalent) |
| `CorsConfig` | struct | `nucleus.go:174` | **unsupported** (`Origins []string` field is dead code; only `AllowAll` is honoured) |
| `New` | func | `nucleus.go:44` | still useful (zero-config entry; `app.New` requires `*Config`) |
| `Load` | func | `nucleus.go:66` | still useful (but panics on error — see §1.2 defects) |
| `CorsAllowAll` | func | `nucleus.go:180` | **unsupported** (constructor for half-implemented type; no callers) |
| `(*AppBuilder).Port` / `.Host` | method | `nucleus.go:83, 89` | **legacy chain method** (one-line config setter) |
| `(*AppBuilder).SQLite` / `.Postgres` / `.MySQL` | method | `nucleus.go:95, 106, 117` | still useful (synthesise URL with sane pool defaults) |
| `(*AppBuilder).WithAdmin` / `.Templates` / `.Static` | method | `nucleus.go:128, 142, 148` | **legacy chain method** |
| `(*AppBuilder).SPA` | method | `nucleus.go:134` | still useful (fallback routing not elsewhere) |
| `(*AppBuilder).Cors` | method | `nucleus.go:155` | **unsupported** (silently dropped — see §1.2 defects) |
| `(*AppBuilder).Provide` | method | `nucleus.go:185` | **unsupported** (`providers` slice never read at `Run()`) |
| `(*AppBuilder).Model` | method | `nucleus.go:191` | still useful (feeds auto-migrate list actually consumed at `Run()`) |
| `(*AppBuilder).AutoMigrate` | method | `nucleus.go:197` | **unsupported** (no-op; migration is triggered by `len(b.models)>0` check) |
| `(*AppBuilder).Run` | method | `nucleus.go:203` | still useful (terminal) |
| `(*AppBuilder).WithConfig` | method | `nucleus.go:270` | still useful (typed escape hatch to `app.Config`) |
| `(*AppBuilder).WithConfigAny` | method | `nucleus.go:277` | **unsupported** (untyped duplicate of `WithConfig`) |
| `(*AppBuilder).Config` | method | `nucleus.go:284` | still useful (read accessor) |
| `(*AppBuilder).Logger` | method | `nucleus.go:289` | still useful (no `app.App` accessor before `Run()`) |
| `Handler` / `Middleware` | type aliases | `routes.go:10, 13` | still useful (handler contract for callers) |
| `RouterGroup` | struct | `routes.go:82` | still useful (only ergonomic prefix+middleware scoping) |
| `(*RouterGroup).Get/Post/Put/Delete` | method | `routes.go:89-154` | still useful |
| `(*AppBuilder).Get/Post/Put/Delete` | method | `routes.go:16-58` | still useful |
| `(*AppBuilder).Group` | method | `routes.go:72` | still useful |
| `(*AppBuilder).Resource` | method | `routes.go:193` | **unsupported** (no callers; uses `:id` path syntax incompatible with the router's `{id}` style) |
| `(*AppBuilder).Use` | method | `routes.go:203` | **unsupported** (range-copy bug — see §1.2 defects) |
| `Resource` (interface) | interface | `routes.go:184` | **unsupported** (no callers; consumer method is broken) |
| `Context` | struct | `context.go:12` | still useful (every handler signature uses `*nucleus.Context`) |
| `(*Context).BindJSON` / `.BindXML` / `.BindForm` | method | `context.go:17-33` | still useful (no `routerpkg.Context` equivalent) |
| `(*Context).Query` / `.Param` | method | `context.go:59, 64` | **legacy chain method** (pass-through to `routerpkg.Context`) |
| `(*Context).JSON` / `.XML` / `.HTML` / `.String` / `.Status` / `.NoContent` / `.Redirect` | method | `context.go:70-109` | mostly **legacy chain method** (thin wrappers); `.XML` may carry value if router-side encoder differs |
| `(*Context).Set` / `.Get` | method | `context.go:116, 121` | **legacy chain method** |
| `(*Context).RequestID` | method | `context.go:126` | still useful (wraps `routerpkg.GetReqID`) |
| `(*Context).SessionGetString` / `.SessionPutString` | method | `context.go:131, 136` | **legacy chain method** (direct delegation) |

Counts: ~22 **still useful**, ~12 **legacy chain method**, ~7
**unsupported**.

### 4.2 Consumer surface map (the entire universe today)

Only two files; both in `examples/ecommerce_dashboard/backend/`.

| Symbol used | Caller : line | Call form | Value semantics? |
|---|---|---|---|
| `New` | `main.go:13` | `nucleus.New()` (chain entry) | held + chained |
| `(*AppBuilder).Port` | `main.go:14` | `.Port(8080)` | chained |
| `(*AppBuilder).SQLite` | `main.go:15` | `.SQLite("ecommerce.db")` | chained |
| `(*AppBuilder).Model` | `main.go:16-20` | `.Model(&models.Product{})` ×4 | chained |
| `(*AppBuilder).AutoMigrate` | `main.go:21` | `.AutoMigrate()` (no-op today) | chained |
| `(*AppBuilder).Group` | `main.go:25` | `app.Group("/api")` → `api` | stored as `*RouterGroup` |
| `(*RouterGroup).Get` / `.Post` | `main.go:26-35` | `api.Get(...)`, `api.Post(...)` | chained on group |
| `SPAConfig` | `main.go:36` | `nucleus.SPAConfig{IndexFile, APIPrefix}` (struct literal) | **field names are observable** |
| `(*AppBuilder).SPA` | `main.go:36` | `.SPA("../frontend/dist", nucleus.SPAConfig{...})` | chained |
| `(*AppBuilder).Run` | `main.go:45` | `app.Run()` (terminal) | returns error |
| `*Context` | `handlers.go:10, 22, 31, 39, 47, 54, 62, 69, 77` | parameter type in every handler | **deepest coupling — every handler signature names this type** |
| `(*Context).JSON` | `handlers.go:19, 25, 29, 43, 50, 65, 73, 80, 83` | `c.JSON(200, ...)` | method call |
| `(*Context).BindJSON` | `handlers.go:33, 57` | `c.BindJSON(&product)` | method call |
| `(*Context).Param` | `handlers.go:40, 70` | `c.Param("id")` | method call |

**`*nucleus.Context` is the deepest coupling.** Every handler in any
real app written against `pkg/nucleus` has `*nucleus.Context` as its
parameter type; v2's answer to this type — keep the `*routerpkg.Context`
embedding, replace it with an interface, make it generic, or expose the
router context directly — determines whether existing handlers compile
unchanged or require mechanical substitution.

### 4.3 Freeze-test impact

**Current scanner state:**
[contracts/freeze_test.go:146-164](contracts/freeze_test.go:146)
hardcodes the package list the baseline scanner walks: `pkg/app`,
`pkg/auth`, `pkg/authz`, `pkg/db`, `pkg/errors`, `pkg/mail`,
`pkg/model`, `pkg/observe`, `pkg/plugins`, `pkg/router`, `pkg/signals`,
`pkg/storage`, `pkg/tasks`, `pkg/validate` — 14 packages. `pkg/nucleus`
is **absent**. Consequence: `TestContractFreeze_APIExportedSymbols_NoRemovals`
cannot fail for `pkg/nucleus` regardless of what is removed or
renamed. This is a false-green situation: the test appears to cover
a stable surface but does not. **The inventory documents this; the
ADR must decide whether to fix it.**

**Per-path impact:**

- **Path (i) — in-place rewrite.** Removing/renaming any current
  `pkg/nucleus` symbol triggers **zero** freeze-test failures today.
  If `pkg/nucleus` were added to the scanner list **before** the
  rewrite, the test would then flag every removed/renamed symbol;
  satisfying it via `NUCLEUS_UPDATE_CONTRACT_BASELINE=1` is
  explicitly prohibited as "hiding a regression" by CLAUDE.md §7.
- **Path (ii) — `pkg/nucleus/v2` coexistence.** No baseline change
  required at any point until the owner deliberately freezes v2.
  Both old and new packages remain outside the scanner.

**Baseline-addition timing** (independent of the path chosen):
recommend deferring inclusion of `pkg/nucleus` (v1 or v2) in
`stableAPISymbolBaselineLines` until the ADR explicitly declares the
v2 API stable. Adding it earlier commits the freeze machinery to a
surface still in flux. `docs/governance/COMPATIBILITY_SLO.md` targets
zero unresolved stable-contract regressions at release; adding an
unstable surface to the baseline manufactures regressions on every
v2 iteration.

**CHANGELOG / DEP- precedent.** ADR-006
([docs/adrs/ADR-006-csrf-hardening.md](docs/adrs/ADR-006-csrf-hardening.md))
and ADR-008
([docs/adrs/ADR-008-csrf-followups.md](docs/adrs/ADR-008-csrf-followups.md))
established the pre-`v1.0` precedent: behaviour changes on stable
surfaces ship as ADR + `BREAKING` CHANGELOG note + one-line migration
per call site, **no `DEP-` entry** because no deprecated-then-removed
cycle is required. The same shape applies here:

- Path (i): every renamed/removed symbol with a caller in
  `examples/` requires a `BREAKING` note; the two consumer files
  ship in the same PR (CLAUDE.md §3 examples rule).
- Path (ii): `Added` entry only when v2 lands; `BREAKING` only at
  the point v1 is removed (if ever); `DEP-` entry appropriate only
  at that removal point.

---

## 5. In-place rewrite vs `pkg/nucleus/v2` — consolidated trade-offs

| Criterion | In-place rewrite of `pkg/nucleus` | Introduce `pkg/nucleus/v2` |
|---|---|---|
| **ADR-001 stdlib-first** | PASS — still delegates to `app.New`, no new dependency | PASS — same; risk only if v2 hides `app.New` behind a new abstraction |
| **Layering** (`pkg/nucleus → pkg/app` direction) | PASS — preserved; dead `b.router` removed | WARN — tree widens; both packages wrap `app.New`; correct but more entry points to document |
| **SPEC §3.1 preservation** | PASS — calls `app.New` as today | PASS — same |
| **Test cycle risk** | Low — no current cycles | Low for v2 tests; moderate if v1 tests later import v2 for comparison |
| **Operating-manual fit (CLAUDE.md §1)** | PASS — directory map already names `pkg/nucleus/` as fluent entry | WARN — directory map and ADR precedence chain must be updated to describe both packages |
| **Migration cost — `ecommerce_dashboard`** | Low: 11 distinct symbols across two files, all in one PR | Higher: same diff plus the import-path swap; still small in absolute terms |
| **Freeze-test pressure today** | None (`pkg/nucleus` absent from scanner) | None |
| **Defect remediation in v1** | Forced — every defect (dead `b.router`, broken `Cors()`, range-copy `Use()`, panic `Load()`, no-op `AutoMigrate()`, plain logger) is fixed by construction | Risk — if v1 is not removed, every defect persists for anyone still importing v1, indefinitely |
| **CHANGELOG shape** | `BREAKING` + ADR for every renamed/removed symbol observable to the example | `Added` only at v2 introduction; `BREAKING` deferred to v1 removal |
| **ADR-007 redaction-aware logger** | Fixed once at the new `New()`; v1 callers no longer affected | Fixed in v2 only; v1 users still bypass redaction unless v1 is also patched |
| **`WithoutDefaults` / `WithExtensions` seam** | Surfaceable in the new chain; old API never had it | Surfaceable in v2; v1 still missing it |

### Recommendation (single paragraph, no decision recorded)

The freeze-test posture today is unusually permissive for
`pkg/nucleus`: the package is entirely absent from the baseline
scanner list at
[contracts/freeze_test.go:146-164](contracts/freeze_test.go:146), so
**either path is mechanically unconstrained by the freeze machinery
today**. That removes the most common argument against an in-place
break and converts the choice into a defect-remediation question and
a migration-window question. The migration cost for the two existing
consumer files is small in absolute terms (eleven distinct symbols in
the chain, plus `*nucleus.Context` as the parameter type in every
handler), but the `*nucleus.Context` coupling is the load-bearing one
for any production consumer that has not yet been written — v2's
answer to that type drives whether real-world handlers compile
unchanged or require a mechanical substitution. Path (i) — in-place
rewrite — forces every defect documented in §1.2 to be fixed by
construction, reaches every `pkg/nucleus` user (today: two example
files) with a single coherent PR, and avoids the indefinite
maintenance burden of two parallel fluent surfaces; its cost is one
BREAKING CHANGELOG entry and an ADR. Path (ii) — `pkg/nucleus/v2`
coexistence — gives the owner a controlled migration window and
lower up-front review surface, at the cost of leaving v1's silent
defects (`Cors()` drop, no-op `AutoMigrate()`, plain-text logger
bypassing ADR-007 redaction) live for whoever still imports it,
indefinitely unless a follow-up removal iteration is scheduled. The
contract-guardian recommendation is the same regardless of path: add
the v2 surface to `stableAPISymbolBaselineLines` in a single commit
coincident with the ADR that declares v2 stable, never freeze the v1
surface, and use the ADR-006/ADR-008 precedent (ADR + `BREAKING`
note, no `DEP-`) for whichever path produces compiler-breaking
changes for the example handlers. The decision belongs to the
upcoming ADR.

---

**End of inventory.** No code changes were made in this pass.
The next `/iterate` cycle picks up from the recommendation above and
opens the ADR (in-place rewrite vs `pkg/nucleus/v2`) before any
implementation begins.
