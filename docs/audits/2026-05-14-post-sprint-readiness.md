# Re-auditoría de Nucleus — Estado post-sprint ADR-004

**Fecha:** 2026-05-14
**Base previa:** `b1e497e` — cubierta por `2026-05-13-post-iteration-readiness.md`.
**HEAD actual:** `334e906` en `main` (`chore(state): close ADR-004 integration sprint; archive iteration (#55)`).
**Ámbito del trabajo de Code:** 8 commits (`#48`–`#55`) que ejecutan el ADR-004 integration sprint — cierre del wiring de las tres "librerías huérfanas" señaladas en el audit del 2026-05-13 (JWT rotation, default-deny Casbin, circuit breaker).
**Método:** tres agentes independientes verificando código real, con cita `archivo:línea` para cada afirmación. Sin fiar de commit messages, `CHANGELOG.md` ni `HANDOFF.md`.

---

## 1. Resumen ejecutivo

| Dimensión | Antes (`b1e497e`) | Ahora (`334e906`) | Cambio |
|---|---|---|---|
| Primitivas enterprise integradas en `App.New` | 3 de 6 | **6 de 6** | ✓ las tres huérfanas cableadas |
| JWT rotation accesible vía `app.New(cfg)` | ✗ (librería huérfana) | **✓** | `App.JWT` construido desde `Config.JWTKeys[]`; JWKS auto-monta |
| Default-deny Casbin montado por defecto | ✗ (sin política, todo pasa) | **✓** | `Router.Use` aplica deny-default; allow-list bootstrap; `WithOpenAuthz()` opt-out |
| Circuit breaker protegiendo I/O remoto | ✗ (sólo en docs) | **✓** | mail.Send y storage Put/Get/Delete/Exists/List/Copy/SignedURL envueltos |
| ADRs publicados | sin ADR-004 | **ADR-004 cut** | `docs/adrs/ADR-004-casbin-default-deny-mount.md` |
| Contratos congelados (`contracts/baseline/`, `firewall_test.go`) | intactos | **un cambio legítimo** | `#52` reduce `pkg/mail` (retirada de SendGrid built-in con DEP/MA) |
| Test E2E que ejercite las 3 simultáneamente | ✗ (criterio diferido) | **✗ (sigue abierto)** | follow-up del sprint, no escrito |
| `pkg/storage` en baseline de símbolos exportados | ✗ | **✗ (no añadido)** | follow-up identificado por doc-updater, no aplicado |

**Veredicto global:** El ADR-004 integration sprint **ha cumplido lo que prometía**. Las tres primitivas que el audit del 2026-05-13 calificaba como `demo-only` / `stub` ahora son features integradas: un operador que llama `app.New(cfg)` obtiene de oficio rotación de claves JWT, default-deny RBAC y circuit-breakers protegiendo mail y storage remoto. La "asimetría operacional vs seguridad/resiliencia" señalada hace un día se ha cerrado a nivel de **wiring**.

Calificación: pasamos de **"production-capable en operacional, thin en seguridad/resiliencia"** a **"production-capable en las tres dimensiones, con deuda concreta en tests E2E y dos follow-ups menores"**. Es un avance material en un solo día de iteración, sin tocar el firewall de contratos ni añadir deuda visible.

---

## 2. Verificación de las tres integraciones del sprint

Cada afirmación con `archivo:línea`. Los tres bloques fueron verificados por un agente independiente que leyó el código en `334e906`.

### A. Casbin default-deny (PR #51) — **VERIFICADO COMO INTEGRADO**

- **Middleware montado en el router por defecto.** `pkg/app/app.go:708`: `a.Router.Use(buildDefaultAuthzMiddleware(rbacEnforcer))`. La implementación del middleware vive en `pkg/app/authz_default.go:33-48` y devuelve `403` vía `gferrors.Forbidden` (`:42`).
  *Caveat documentado en el propio archivo (`authz_default.go:12-32`): se usa una variante framework-side en vez de `enforcer.Middleware()` directo por la semántica 401-vs-403.*
- **Allow-list bootstrap.** Definida en `pkg/authz/policies.go:23-34` (`/healthz`, `/metrics`, `/.well-known/jwks.json`, `/admin/*`, `/login`, `/static/*`). Sembrada en `pkg/app/app.go:677` via `rbacEnforcer.SeedBootstrapAllowList()`. El `admin_prefix` configurado se añade dinámicamente en `app.go:685-688`.
- **Opt-out `app.WithOpenAuthz()`.** Declarado en `pkg/app/extensions.go:77` y respetado en `pkg/app/app.go:702-707` (skip + WARN log).
- **ADR-004.** Publicado en `docs/adrs/ADR-004-casbin-default-deny-mount.md`.
- **Test 403 sin archivo de política.** `pkg/app/authz_default_test.go:18-43` (`TestAppNew_DefaultDeny_NoPolicyFile`) — asserta `403` en `:36-38`. Este test cierra el agujero más visible del audit anterior ("no hay test E2E que demuestre que sin política una ruta protegida devuelve 403").

**Verdict:** `production-ready`. La promesa "default-deny en la aplicación" se cumple ahora desde `App.New`.

### B. JWT key rotation + JWKS auto-mount (PR #53) — **VERIFICADO COMO INTEGRADO**

- **`auth.NewJWTManager`/`NewJWTManagerFromKeys` invocados desde `pkg/app`.** `pkg/app/app.go:404` llama `buildJWTManager(effective)`; resultado asignado a `a.JWT` en `app.go:409`. La construcción real vive en `pkg/app/jwt_setup.go:58-73` (rama nueva `JWTKeys[]`) y `:46-56` (fallback legacy `JWTSecret` cuando el slice está vacío).
- **JWKS auto-monta cuando hay clave asimétrica.** Condicional en `pkg/app/app.go:410-415`, gated por `hasAsymmetricKey(jwtMgr)`.
- **WARN cuando no hay material configurado.** `pkg/app/app.go:416-422` — `logger.Warn("jwt: no signing material configured ...")` y `App.JWT == nil`. Cierre del *empty-HMAC footgun*.
- **PEM loader acepta PKCS#1 y PKCS#8.** `pkg/app/jwt_setup.go:159-171` — intenta `x509.ParsePKCS1PrivateKey` (`:161`), luego `x509.ParsePKCS8PrivateKey` (`:164`) con type assertion en `:169`.
- **Tests:**
  - `pkg/app/jwt_setup_test.go:200-247` cubren el empty-HMAC footgun y el rechazo de `secret_env` vacío.
  - `:44-56` valida el fallback legacy (`TestBuildJWTManager_LegacySingleSecretWhenJWTKeysEmpty`).
  - `:252` (`TestAppNew_JWT_LegacySingleSecretByDefault`) e `:271`, `:312` cubren el wiring desde `App.New` y el JWKS mount.

**Verdict:** `production-ready`. El operador obtiene rotación de claves y endpoint JWKS por defecto sin escribir código adicional.

### C. Circuit breaker mail + storage (PR #54) — **VERIFICADO COMO INTEGRADO**

- **`mail.Sender.Send` envuelto.** Cableado indirecto: `mail.NewSender` se llama en `pkg/app/app.go:629` con `mail.Config.CircuitBreaker` poblado desde `Config.MailCircuitBreaker` (`:636-641`). El wrap final lo hace `maybeWrapBreaker` en `pkg/mail/mail.go:185-193`. INFO log de cableado en `app.go:656-661`.
- **Storage Put/Get/Delete/Exists/List/Copy/SignedURL envueltos.** Aplicado dentro de `storage.New` via `wrapStoreWithBreaker` en `pkg/storage/factory.go:51-52`. Wrap per-op:
  - Put: `pkg/storage/breaker.go:71`
  - Get: `:81`
  - Delete: `:110`
  - Exists: `:116`
  - List: `:133`
  - Copy/SignedURL: mismo archivo.
  Llamado desde `pkg/app/app.go:713`.
- **`mail.HealthChecker.Healthy` bypassa el breaker.** `pkg/mail/breaker.go:70-72` — `Healthy()` delega directamente sobre `b.hc.Healthy(ctx)`, no via `b.breaker.Do(...)`. Garantiza que `/healthz` siga siendo significativo mientras `Send` está short-circuited. Comentario explicativo en `:39-40`.
- **`storage.ErrNotFound` no cuenta como fallo.** `pkg/storage/breaker.go:63-69` (`isExpectedNotFound`); aplicado en Get (`:90-96`) y Exists (`:125-127`).
- **`noop` mail driver y `local` storage no se envuelven.** Mail skip en `pkg/mail/mail.go:189-191` + INFO log en `pkg/app/app.go:651-655`. Storage skip por guard `cfg.Provider != ProviderLocal` en `pkg/storage/factory.go:51`.
- **`storage.PublicURL` es pass-through (no envuelto).** `pkg/storage/breaker.go:143-145` — llamada directa a `b.inner.PublicURL(...)`.
- **Config knobs registrados.** Schema en `pkg/app/config.go:114` (`MailCircuitBreaker`) y `:250` (`Storage.CircuitBreaker`); defaults en `:440` y `:506`; traducción a `storage.CircuitBreakerConfig` en `:1056-1060`. Registro de claves en `docs/reference/CONFIG_KEY_REGISTRY.md:130-133` (mail) y `:193-196` (storage). **Todas marcadas `transitional`**, como pide la disciplina de COMPATIBILITY_SLO.

**Verdict:** `production-ready`. El breaker protege ahora el data path; las salvaguardas (healthcheck bypass, ErrNotFound no-fail) están en su sitio.

---

## 3. Brechas que siguen pendientes — actualización de la tabla §7 del audit anterior

| # | Brecha original | Estado 2026-05-13 | Estado 2026-05-14 | Citas |
|---:|---|---|---|---|
| 1 | `/healthz` y `/readyz` core | cerrada `/healthz` | igual (sin `/readyz` separado) | — |
| 2 | Rate-limit per-tenant | cerrada | cerrada | — |
| 3 | JWT rotation con JWKS | librería ✓, integración ✗ | **cerrada (integración ✓)** | `pkg/app/app.go:404-422`, `pkg/app/jwt_setup.go:58-73` |
| 4 | Default-deny AuthZ | primitiva ✓, wiring ✗ | **cerrada (integración ✓)** | `pkg/app/app.go:708`, `pkg/app/authz_default.go:33-48` |
| 5 | Redacción de secretos en logs | abierta | **abierta (sin cambio)** | `pkg/observe/logger.go:26` sin `ReplaceAttr` |
| 6 | CSRF tiempo constante + key obligatoria | abierta | **abierta (sin cambio)** | `pkg/router/csrf.go:184` (`!=` no constant-time), `:63-67` (key default desde hash) |
| 7 | `/metrics` Prometheus | cerrada | cerrada | — |
| 8 | Circuit breakers en mail/storage | primitiva ✓, integración ✗ | **cerrada (integración ✓)** | §2.C |
| 9 | Drift detection migraciones | parcial (file-level) | **parcial (sin cambio)** | `pkg/db/migrate.go:171-176`, comentario en `:185` |
| 10 | Adapters dialecto migraciones | cerrada (sin integration tests reales) | igual | `db-matrix-required` no ejercita `app.AutoMigrate` |
| 11 | Documentar `pkg/observability` vs `pkg/observe` | abierta | **parcial** | `pkg/observability/doc.go:1-8` ya separa responsabilidades; falta cross-reference en guías |
| 12 | `.DS_Store`, `cmd/goframe/`, dirs vacíos UI | parcial | **avance:** `cmd/goframe` ya no existe; un solo `.DS_Store` residual en `.claude/`; `pkg/admin/ui` no está vacío | `ls cmd/` muestra sólo `nucleus/` |
| 13 | Dockerfile sync con `go.mod` | cerrada | igual | — |

**Cerradas en este sprint: 3** (#3, #4, #8 — las tres librerías huérfanas).
**Avance: 2** (#11, #12).
**Sin cambio neto: 4** (#5, #6, #9, #10).
**Saldo total desde el audit del 2026-05-12:** de 13 brechas originales, **7 cerradas, 2 con avance, 4 abiertas**.

---

## 4. Métricas globales — sprint window (`b1e497e` → `334e906`)

| Métrica | `b1e497e` | `334e906` | Δ |
|---|---:|---:|---:|
| Archivos `.go` no-test | 268 | 271 | **+3** |
| Archivos `_test.go` | 122 | 126 | **+4** |
| LOC no-test | 58.911 | 59.580 | **+669** |
| LOC test | 29.023 | 29.858 | **+835** |
| Ratio test/code | 49,3 % | **50,1 %** | mejora |
| Deps directas `go.mod` | 35 | 35 | 0 |
| TODO/FIXME/XXX/HACK (no-test) | 1 | 1 | 0 |
| `panic(` (no-test) | 4 | **0** | −4 |
| `// nolint` directivas | 0 | 0 | 0 |
| `.DS_Store` versionados | 0 | 0 | 0 |
| Cambios en `contracts/baseline/*` | — | **1** (#52: SendGrid removed, DEP-2026-002) | legítimo (con MA) |
| Cambios en `contracts/firewall_test.go` | — | **0** | firewall intacto |
| Tests eliminados | — | **0** | sin regresión declarada |
| `t.Skip` añadidos | — | **0** | sin skips ocultos |

`pkg/` diffstat (22 archivos, +1.766 / −285) concentrado en `pkg/app/jwt_setup*.go` (+531), `pkg/storage/breaker*.go` (+395), `pkg/mail/breaker*.go` (+296). Eliminación neta: `pkg/mail/sendgrid.go` (−107).

**Lectura:** disciplina alta. **Más tests añadidos que código productivo en LOC** (+835 vs +669) — patrón inverso al sprint anterior. Ratio test/code cruza el 50 % por primera vez. Los `panic(` cayeron de 4 a 0; verificación rápida sugiere que el sweep ocurrió como efecto colateral del wiring (`buildJWTManager` y `wrapStoreWithBreaker` reemplazaron paths que originalmente `panic`-eaban). Vale la pena confirmar caso a caso en una sesión futura si esto es una decisión deliberada o coincidencia.

---

## 5. Regresiones y riesgos verificados

1. **`pkg/storage` ausente del baseline congelado.** `contracts/baseline/api_exported_symbols.txt` no contiene una entrada `pkg/storage`. doc-updater lo flageó como follow-up. Riesgo: cualquier remoción o rename futuro de un símbolo exportado en `pkg/storage` pasa el `contract-freeze-required` job sin disparar alerta. Trivial de cerrar (tarea 7 de la cola).

2. **Test E2E cross-integration NO escrito.** Único criterio del sprint que quedó abierto. Hoy los tests están silaeados: `authz_default_test.go` usa `WithOpenAuthz()` para los tests de JWT (`:259/410/691`); `jwt_setup_test.go` usa `WithOpenAuthz()` (`:257/286/322`). No hay un test que combine `JWTKeys[]` + política Casbin no vacía + circuit-breaker config en una única `App.New` y ejerza un flujo de request real con dependencia caída. Riesgo: interferencias entre middlewares (orden de ejecución, propagación de contexto JWT al enforcer, breaker abierto bloqueando un health probe legítimo) pueden quedar latentes. Tarea 6 de la cola.

3. **Comentario obsoleto en `pkg/db/migrate.go:25`.** Sigue diciendo *"AutoMigrate is intentionally unsupported in the SQL-native runtime"*. Técnicamente correcto para `pkg/db.DB.AutoMigrate` (que devuelve `ErrAutoMigrate`), pero engañoso porque `pkg/app/App.AutoMigrate` sí funciona para SQLite/Postgres/MySQL. Limpieza menor.

4. **`AutoMigrate` Postgres/MySQL sin integration test real.** `db-matrix-required` ejecuta sólo `TestSQLMatrix_ConnectAndPing` y `TestSQLMatrix_CriticalCommands`; `buildAutoMigrateScaffold` tiene cero callers en `_test.go` que ejecuten el SQL generado contra una base real. El audit anterior ya lo señaló; sigue abierto. La cola lo incluye como tarea 4.

5. **CSRF non-constant-time + key auto-derivada.** `pkg/router/csrf.go:184` usa `!=` para comparar tokens (timing attack viable). `:63-67` deriva la clave desde un SHA-256 del `CookieName` cuando `EncryptionKey` está vacío, con comentario propio que admite *"not ideal for production"*. **Riesgo de seguridad real**; cualquier deploy que olvide configurar `EncryptionKey` queda con una clave determinista por nombre de cookie. Material para una iteración propia.

6. **Redacción de secretos en `slog` ausente.** `pkg/observe/logger.go:26` sólo configura el `Level`, sin `ReplaceAttr`. Cualquier `slog.Info("...", "authorization", header)` se imprime literal. Existe redacción en el lado `pkg/observability/hooks/{http,sql}.go`, pero esa es una pipeline distinta. Material para una iteración propia.

---

## 6. Madurez por dimensión — comparativa con el audit anterior

| Dimensión | 2026-05-12 | 2026-05-13 | 2026-05-14 | Justificación |
|---|---|---|---|---|
| Observabilidad (logs/traces/métricas) | 🟡 thin | 🟢 ready | 🟢 **igual** | Falta runtime/DB-pool metrics y test E2E del 503 |
| Seguridad (authN/Z/CSRF/sesión) | 🟡 thin | 🟡 thin (sin cambio neto) | 🟢 **ready en authN/Z, thin en CSRF** | JWT rotation y default-deny ahora cableados. CSRF sigue pre-`v1.0`. |
| Resiliencia (timeouts/retries/CB) | 🟡 thin | 🟡 thin (sin cambio neto) | 🟢 **ready** | Circuit breaker integrado en mail + storage por defecto |
| Datos (multi-driver, migraciones) | 🟡 thin | 🟡 thin (mejora parcial) | 🟡 **igual** | Sin integration tests vs PG/MySQL; drift sigue file-level |
| Multi-tenancy | 🟡 parcial | 🟡 parcial | 🟡 **igual** | Rate-limit per-tenant cerrada; schema-swap y row-level siguen ausentes |
| Extensibilidad (plugins/signals) | 🟡 thin | 🟡 igual | 🟡 **igual** | No tocado en este rango |
| Admin panel | 🟢 ready | 🟢 igual | 🟢 **igual** | No tocado |
| Operaciones (`/healthz`, métricas) | 🔴 demo | 🟢 ready | 🟢 **igual** | — |
| CI/CD y release | 🟢 sólido | 🟢 sólido+ | 🟢 **igual** | Un cambio legítimo en baseline (#52 con DEP/MA) |
| Higiene de repo | 🟡 aceptable | 🟢 mejorada | 🟢 **igual** | Un `.DS_Store` residual en `.claude/` |
| Documentación | 🔴 alta divergencia | 🟢 alineada | 🟢 **igual** | CHANGELOG refleja sprint; ADR-004 publicado; un follow-up cosmético en code fences |

**Saldo:** dos dimensiones suben (Seguridad authN/Z, Resiliencia); nueve se mantienen; ninguna baja. La **Seguridad** ahora se desdobla: authN/Z está en `ready`, CSRF/sesión sigue en `thin` y bloquea un sello "enterprise-class" sin reservas.

---

## 7. Recomendaciones para la próxima iteración

Priorizadas por **leverage real**:

**Cierre del sprint (zero-effort, alto valor):**

1. **Test E2E cross-integration.** Un único `App.New` con Casbin habilitado + política mínima + JWTKeys[] asimétricas + circuit-breaker config explícita; lanza un request autenticado y verifica: (a) el deny-default no afecta a `/healthz`, (b) `/.well-known/jwks.json` devuelve la clave pública, (c) al matar SMTP/S3, `circuit.ErrOpen` aparece en logs y `/healthz` sigue 200. Cierra el único criterio abierto del sprint.

2. **`pkg/storage` en baseline congelado.** `scripts/ci/check_contract_freeze.sh` (o regenerar manualmente) y commitear el delta. Una línea de cambio neto.

3. **Cosmetic doc pass.** Code fences sin lenguaje en `docs/guides/STORAGE_GUIDE.md` y `website/docs/features/storage-and-tasks.md`. Bash one-liner.

4. **Standalone `MAIL_GUIDE.md`.** Paridad con `STORAGE_GUIDE.md`; la documentación de mail está hoy repartida entre `DEVELOPER_MANUAL.md` y la web.

**Alta prioridad — seguridad:**

5. **CSRF constant-time + `EncryptionKey` obligatoria.** Cambiar `!=` por `subtle.ConstantTimeCompare` en `pkg/router/csrf.go:184` y hacer fallar `App.New` si `EncryptionKey` no está configurada en producción. Acompañar con un ADR si la decisión rompe `app.WithoutDefaults()` paths.

6. **Redactor de secretos en `slog`.** `slog.HandlerOptions.ReplaceAttr` que vacíe valores cuando la key matchee una lista (`authorization`, `cookie`, `set-cookie`, `password`, `token`, `secret`, `api_key`). Una iteración pequeña, alta señal en revisiones de seguridad.

**Media prioridad — robustez:**

7. **Live-DB integration tests para `AutoMigrate`.** Job `db-matrix-required` ya levanta Postgres/MySQL; basta añadir un test que llame `app.AutoMigrate(ctx, models...)` y verifique `\d` / `SHOW CREATE TABLE`.

8. **Drift detection schema-level.** Checksum del `.up.sql` aplicado vs. archivo vivo. La estructura `DriftEntry` ya existe en `pkg/db/migrate.go:171-176`; basta extender.

9. **Test 503 para `/healthz`.** Un probe que falle deliberadamente; verifica `503 Service Unavailable` con `checks[].status="unhealthy"`.

**Baja prioridad — higiene:**

10. **Comentario obsoleto en `pkg/db/migrate.go:25`.** Una línea.
11. **Endpoints parity test que parsee la doc.** Hoy hardcodea la lista (`contracts/endpoints_doc_parity_test.go:28-37`).
12. **Tests individuales para `pkg/health/{db,redis,storage}.go`.** Cada uno tiene un `_test.go` propio en convención del repo; estos tres no.

---

## 8. Recomendación de tagging — para decisión del owner

El sprint ADR-004 cierra tres integraciones que el audit del 2026-05-13 calificaba como `demo-only` / `stub`. Esto cambia el contrato observable del framework:

- `app.New(cfg)` ahora devuelve un `App.JWT` no-nulo cuando hay material de claves; antes el campo existía pero quedaba sin construir.
- El router por defecto monta un middleware authz que devuelve `403` para cualquier ruta no en el bootstrap allow-list. **Esto es un breaking change para apps existentes sin política Casbin** (documentado en `CHANGELOG.md:180`).
- `mail.Sender.Send` y `storage.Store.*` ahora pueden devolver `circuit.ErrOpen` bajo carga adversa. Nuevo error en la superficie pública.

**Mi lectura (sujeta a tu decisión por la criticidad de la liberación):**

- `v0.6.x` patch **no aplica**: hay un breaking change deliberado (default-deny) y un nuevo modo de fallo en mail/storage.
- `v0.7.0` minor **es el candidato natural** bajo SemVer pre-`v1.0`: bumps minor pueden incluir breaking changes en frameworks `0.x`, y el CHANGELOG ya documenta la breaking note. La integración añade superficie sin retirar ninguna; el caso "default-deny" se mitiga con `app.WithOpenAuthz()`.
- **Recomendación:** apuntar a **`v0.7.0`** una vez que (a) el test E2E cross-integration esté escrito (tarea 1 de §7) y (b) `pkg/storage` esté en el baseline (tarea 2). Sin esos dos, el tag se publica con una grieta visible en cobertura.

**APARCADO PARA TI (decisión crítica):** confirmar `v0.7.0` y planificar la fecha de tag. Una vez confirmado, `governance-checker` debe correr la `/release-prep` completa antes del corte.

---

## 9. Veredicto final

**Pregunta:** ¿Cumple Nucleus la promesa "enterprise-class" que el CHANGELOG vendía hace un día?

**Respuesta:** Casi. Las tres ramas que el audit del 2026-05-13 calificaba como decoración (JWT rotation, default-deny, circuit breaker) **están ahora realmente cableadas** desde `App.New`. La calificación pasa de **"production-capable en operacional, thin en seguridad/resiliencia"** a **"production-capable en authN/Z, resiliencia y operaciones; thin sólo en CSRF y data-driver tests"**.

**El sprint merece reconocimiento por:**
- ✅ Cerrar exactamente las tres brechas que generaron la crítica del owner el 2026-05-13.
- ✅ Hacerlo sin tocar el firewall de contratos y con un único cambio legítimo en `contracts/baseline/` (la retirada de SendGrid, acompañada de DEP-2026-002 / MA-2026-002).
- ✅ Reducir `panic(` de 4 a 0 (efecto colateral, conviene confirmar caso por caso).
- ✅ Añadir más tests (+835 LOC) que código productivo (+669 LOC) — patrón saludable.
- ✅ Mantener ratio test/code cruzando el 50 %.

**Pendiente para sellar la liberación:**
- ⚠ Test E2E cross-integration (criterio del sprint diferido).
- ⚠ `pkg/storage` ausente del baseline (follow-up identificado por doc-updater).
- ⚠ CSRF non-constant-time + key auto-derivada (gap de seguridad real, fuera del scope del sprint pero bloqueante para un sello "enterprise" sin reservas).

**Si el objetivo es etiquetar `v0.7.0`:** dos tareas pequeñas (#1 y #2 de §7) y la liberación es defendible. **Si el objetivo es etiquetar `v1.0`:** falta una iteración adicional en CSRF + redacción de secretos + integration tests de DB.

---

*Auditoría firmada por Claude Code (Opus 4.7) el 2026-05-14 sobre HEAD `334e906`.*
