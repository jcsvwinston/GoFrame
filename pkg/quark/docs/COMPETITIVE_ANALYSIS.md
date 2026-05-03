# Quark ORM — Análisis Competitivo Exhaustivo y Hoja de Ruta

> **Fecha**: Mayo 2026  
> **Propósito**: Auditoría técnica completa, verificación implementación vs documentación, análisis de cobertura de tests, y comparativa real con ORMs competidores en Go.  
> **Audiencia**: Equipo de desarrollo, previo a la publicación como paquete independiente.

---

## PARTE 1: AUDITORÍA INTERNA — IMPLEMENTACIÓN vs DOCUMENTACIÓN

### 1.1 Funcionalidades Verificadas como REALMENTE IMPLEMENTADAS ✅

| Feature | Archivo(s) | Tests | Estado |
|---------|-----------|-------|--------|
| Generic Query Builder `For[T]` | `client.go:112-166` | `quark_test.go`, `suite_test.go` | ✅ Completo |
| CRUD: Create, Find, First, List | `query_crud.go`, `query_exec.go` | `quark_test.go:61-183` | ✅ Completo |
| Update (entity + Map) | `query_crud.go:392-456` | `quark_test.go:147-225` | ✅ Completo |
| Delete (Soft + Hard + DeleteBy) | `query_crud.go:640-898` | `quark_test.go:227-293` | ✅ Completo |
| Where / WhereIn / WhereBetween / Or | `query_builder.go:113-182` | `suite_test.go:253-320`, `features_test.go:293-321` | ✅ Completo |
| WhereJSON (cross-dialect) | `query_builder.go:252-263` | `suite_test.go:796-833` | ✅ Completo |
| OrderBy / Limit / Offset | `query_builder.go:184-207` | `quark_test.go:446-484` | ✅ Completo |
| Join / LeftJoin / RightJoin | `query_builder.go:216-234` | `features_test.go:240-289` | ✅ Completo |
| Select (columnas específicas) | `query_builder.go:106-110` | `suite_test.go:311-319` | ✅ Completo |
| Pagination (`Paginate`) | `page.go` | `quark_test.go:366-404`, `suite_test.go:431-454` | ✅ Completo |
| Streaming `Iter()` | `query_exec.go:246-288` | `quark_test.go:296-324` | ✅ Completo |
| Cursor manual | `cursor.go` | `quark_test.go:326-364` | ✅ Completo |
| Count | `query_exec.go:291-348` | `quark_test.go:407-444` | ✅ Completo |
| Transacciones (Callback + Manual) | `tx.go` | `quark_test.go:486-539`, `features_test.go:48-171` | ✅ Completo |
| Savepoints + Nested Tx | `tx.go:86-134` | `features_test.go:129-171`, `quark_test.go:486-539` | ✅ Completo |
| Soft Delete (deleted_at) | `query_crud.go:640-803` | `suite_test.go:395-429` | ✅ Completo |
| Unscoped (incluir soft-deleted) | `query_builder.go:98-103` | `suite_test.go:422-429` | ✅ Completo |
| Hooks (Before/After Create/Update/Delete) | `hooks.go` | `quark_test.go:621-708` | ✅ Completo |
| Validación (validator/v10 + interfaz) | `validator.go` | `quark_test.go:710-766`, `suite_test.go:380-393` | ✅ Completo |
| 6 Dialectos (PG, MySQL, SQLite, MSSQL, Oracle, MariaDB) | `dialect.go` | `dialect_test.go` | ✅ Completo |
| Custom Dialect Registry | `dialect.go:766-817` | `dialect_test.go:191-216` | ✅ Completo |
| Auto-detect dialect | `client.go:257-290` | `dialect_test.go:191-216` | ✅ Completo |
| SQLGuard (SQL Injection Prevention) | `internal/guard/guard.go` | Validaciones integradas en todos los builders | ✅ Completo |
| Model Metadata Cache (sync.Map, O(1)) | `internal/schema/schema.go` | Implícito en todos los tests | ✅ Completo |
| Immutable Query Clone | `query_builder.go:77-89` | `features_test.go:175-229` | ✅ Completo |
| Eager Loading (Preload) — has_one, has_many, belongs_to | `query_exec.go:604-985` | `quark_test.go:553-619`, `suite_test.go:747-794` | ✅ Completo |
| M2M Relations (many_to_many) | `query_exec.go:728-872` | `suite_test.go` (implicit) | ✅ Completo |
| Polymorphic Relations | `query_exec.go:874-935` | Schema parsing verified | ✅ Parcial (sin test E2E dedicado) |
| Recursive Association Saving | `query_crud.go:112-259, 900-966` | `suite_test.go:747-794`, `association_test.go` | ✅ Completo |
| M2M Link in Join Table | `query_crud.go:968-984` | `suite_test.go` (implicit) | ✅ Completo |
| Composite PK Support | `query_crud.go`, `internal/schema/schema.go` | `composite_pk_test.go` (356 líneas) | ✅ Completo |
| Auto-Migration (CREATE TABLE) | `migrator.go` | `quark_test.go:716-766`, `suite_test.go` | ✅ Completo |
| Evolutionary Sync (Add/Rename/Drop columns) | `sync.go` | `suite_test.go:701-745` | ✅ Completo |
| Versioned Migrations (Up/Down) | `migrate/migrate.go` | Solo CLI, sin test unitario | ⚠️ Sin tests directos |
| Multi-Tenant: RowLevelSecurity | `tenant_router.go`, `client.go:143-163` | `suite_test.go:456-497`, `quark_test.go:768-801` | ✅ Completo |
| Multi-Tenant: SchemaPerTenant | `tenant_router.go`, `client.go:150-151` | `quark_test.go:797-801` | ✅ Básico |
| Multi-Tenant: DatabasePerTenant (LRU) | `tenant_router.go:112-159` | `suite_test.go:602-644` | ✅ Completo |
| Middleware Pipeline | `option.go:109-125`, `query_crud.go:48-53`, `query_exec.go:24-29` | `features_test.go:323-387`, `suite_test.go:560-574` | ✅ Completo |
| Query Observers | `option.go:83-98` | `suite_test.go:499-539` | ✅ Completo |
| Cache Layer (Memory + Redis) | `cache.go`, `cache/memory/`, `cache/redis/` | `suite_test.go:112-185` | ✅ Completo |
| Cache Tag-Based Invalidation | `query_crud.go:26-28`, `cache/memory/memory.go:76-89` | `suite_test.go:170-185` | ✅ Completo |
| OpenTelemetry Middleware | `otel/otel.go` | `otel_test.go` (460 líneas) | ✅ Completo |
| Routine Builder (Functions/Procedures) | `routine_builder.go` | `quark_test.go:803-818` | ✅ Básico |
| Call (Stored Procedures) | `routine_builder.go:125-139` | `quark_test.go:803-818` | ✅ Básico |
| EventBus / Notify | `events.go` | `quark_test.go:820-831` | ⚠️ Stub (no implementado realmente) |
| RawQuery / Exec | `client.go:178-239` | `suite_test.go:576-600` | ✅ Completo |
| CLI (init, model, migrate, seed, sync, inspect, tenant, validate) | `cli/commands/` | Sin tests de CLI | ⚠️ Sin tests |
| Benchmark Engine Test | `benchmark_test.go` | Manual (testing.Short skip) | ✅ Completo |
| Stress Test | `stress_test.go` | 372 líneas | ✅ Completo |

### 1.2 Discrepancias Documentación vs Implementación 🔴

| Claim en README/Docs | Realidad | Severidad |
|----------------------|----------|-----------|
| **EventBus / LISTEN-NOTIFY**: "Quark introduce el EventBus" | `events.go:43-55`: Devuelve error hardcoded "not yet mapped in V1". **NO funciona para ningún dialecto.** | 🔴 ALTA |
| **Notify**: Documentado como funcional | Solo funciona en Postgres (via `pg_notify`), para otros dialectos devuelve error | 🟡 MEDIA |
| **Routine.List()**: Documentado como soporte genérico | `routine_builder.go:50`: Bypasses middleware. No usa `executeQuery` ni `executeExec`. | 🟡 MEDIA |
| **Call con OUT params**: `sql.Out{Dest: &procesados}` | Funciona pero sin test real con OUT params | 🟡 MEDIA |
| **"Update() actualiza todos los campos"** | `query_crud.go:493-494`: `isZeroValue` skip — es **partial update** (zero values se omiten). README dice "actualiza todos los campos" pero es falso. | 🔴 ALTA |
| **Release Notes v1.0.0**: "production-ready" | EventBus es stub, Migrator usa `?` hardcoded (no dialect-aware), varios edge cases sin cubrir | 🟡 MEDIA |
| **Introspection MariaDB**: No existe handler en `introspection.go` | `GetTableInfo` no tiene case para "mariadb", fallará con "unsupported dialect" | 🔴 ALTA |

### 1.3 Bugs y Problemas Técnicos Detectados 🐛

1. **`migrate/migrate.go:97`**: Usa `?` hardcoded como placeholder — incompatible con PostgreSQL (`$1`), MSSQL (`@p1`), Oracle (`:1`). El Migrator versioned solo funciona con MySQL/SQLite.

2. **`migrate/migrate.go:187`**: Mismo problema con `DELETE FROM ... WHERE id = ?`.

3. **`features_test.go:353-387` (TestMiddlewareChain)**: Los middlewares `m1` y `m2` nunca se conectan al client (el test crea el client sin ellos). Las assertions sobre `m1.execs` y `m2.execs` siempre fallarán porque los middlewares nunca se invocan. **Test roto.**

4. **`query_exec.go:605-607`**: `loadRelations` modifica `AllowRawQueries` temporalmente — **race condition** si se ejecuta concurrentemente con otras queries.

5. **`tenant_router.go:124`**: Factory se ejecuta bajo lock (`r.mu`). Si la factory es lenta (red), bloquea a todos los demás tenants. El comentario reconoce el problema pero no hay `singleflight`.

6. **`memory.go:91-103`**: `cleanupLoop` goroutine nunca se detiene. No hay mecanismo de `Close()` para el Store. **Goroutine leak** en tests.

7. **`redis.go:43`**: El prefijo de clave es `quark:cache:` + key, pero `generateCacheKey` ya produce claves con prefijo `quark:cache:`. Resultado: claves Redis con **doble prefijo** `quark:cache:quark:cache:...`.

8. **`Paginate`** (`page.go:37`): Muta `q.limit` y `q.offset` directamente **sin clonar** — viola el contrato de inmutabilidad del Query builder.

---

## PARTE 2: ANÁLISIS DE COBERTURA DE TESTS

### 2.1 Funcionalidades CON tests adecuados ✅

- CRUD completo (Create, Find, First, List, Update, UpdateMap, Delete, HardDelete, DeleteBy)
- Query Builder (Where, WhereIn, WhereBetween, Or, OrderBy, Limit, Offset, Select)
- Transacciones (commit, rollback, manual, nested/savepoints)
- Soft Delete + Unscoped
- Hooks (Before/After Create/Update/Delete)
- Validación (validator/v10 + interfaz)
- Immutable Query Clone + Concurrent Safety
- JOINs (Inner, Left)
- Pagination
- Streaming (Iter) + Cursor
- Count
- Eager Loading (has_many, belongs_to, has_one)
- Composite PK (Create, Update, HardDelete, DeleteBy, List, schema detection, string keys)
- Multi-Tenant RLS (isolation, auto-inject tenant_id)
- Multi-Tenant DatabasePerTenant (LRU, eviction)
- Cache (Memory + Redis, invalidation)
- OpenTelemetry (spans for CRUD, attributes, transactions, context propagation)
- Dialect tests (MSSQL, Oracle, MariaDB, detection)
- Sync (add column, rename column, destructive drop)
- Recursive Association Saving + Preload
- Benchmark + Stress test
- Cross-engine SharedSuite (SQLite, PostgreSQL, MySQL, MSSQL, Oracle, MariaDB)

### 2.2 Funcionalidades SIN tests o con cobertura insuficiente ❌

| Feature | Gap |
|---------|-----|
| **Polymorphic Relations** | Solo se parsea el tag en schema.go. No hay test E2E de Preload polymorphic. |
| **M2M Preload** | No hay test dedicado que verifique `Preload("Roles")` en un modelo M2M con join table. |
| **Versioned Migrations (migrate/migrate.go)** | Cero tests unitarios. Up/Down/DryRun completamente sin cobertura. |
| **EventBus / EventListener** | Solo test de error (SQLite no soporta). Ningún test funcional. |
| **Routine[T].List()/First()/Scalar()** | Test mínimo (`quark_test.go:803-818`): solo `Call` con SQLite `abs`. No testa `List()` ni `First()`. |
| **CLI commands** | Cero tests para init, model, migrate, seed, sync, inspect, tenant, validate. |
| **RightJoin** | No hay test. Solo Left e Inner. |
| **WhereJSON** | Test existe pero skippea Oracle, MSSQL, MariaDB. Solo PostgreSQL/SQLite realmente probado. |
| **Cache Redis** | Skip si Redis no disponible. Nunca se ejecuta en CI local típico. |
| **SchemaPerTenant** | Solo verifica que no hay error al crear query, no verifica que el SQL generado incluye schema prefix. |
| **Introspection** | Sin tests unitarios para `internal/db/introspection.go`. |
| **SQLGuard** | Sin tests unitarios para `internal/guard/guard.go`. |
| **DryRun Sync** | Se invoca en tests pero no se verifica que NO ejecute SQL. |
| **Custom TableName** | Implícito en SyncUserV1-V4, pero sin test explícito del interface `TableNamer`. |
| **Middleware WrapQueryRow** | Ningún test verifica que WrapQueryRow sea invocado correctamente. |
| **Limits enforcement** | `MaxWhereConditions` y `MaxQueryLength` definidos pero **nunca verificados** en código ni tests. |
| **Error wrapping** | `ErrTimeout`, `ErrConstraintViolation` definidos pero nunca utilizados en código. |
| **Cache Tenant-Aware** | Documentado ("Cache keys include Tenant ID") pero nunca testeado con multi-tenant + cache. |

### 2.3 Estimación de Cobertura

| Área | Cobertura Estimada |
|------|--------------------|
| Core CRUD | ~95% |
| Query Builder | ~90% |
| Transactions | ~95% |
| Relations (standard) | ~85% |
| Relations (M2M + Poly) | ~40% |
| Multi-Tenant | ~70% |
| Cache | ~75% |
| OTel | ~90% |
| Migrations (auto) | ~80% |
| Migrations (versioned) | ~0% |
| CLI | ~0% |
| Security (Guard) | ~60% (integración, no unitarios) |
| Introspection | ~0% (unitarios) |
| Events | ~5% |
| Routines | ~20% |
| **GLOBAL ESTIMADA** | **~65-70%** |

---

## PARTE 3: COMPARATIVA CON ORMs COMPETIDORES EN Go

### 3.1 Competidores Analizados

| ORM | GitHub Stars | Madurez | Enfoque |
|-----|-------------|---------|---------|
| **GORM** | ~37k | Muy maduro (2013+) | Full-featured, convention-over-config |
| **Ent** (Meta/Facebook) | ~16k | Maduro (2019+) | Code-generation, graph-based schema |
| **SQLBoiler** | ~7k | Maduro (2016+) | Code-gen from DB, type-safe |
| **Bun** | ~4k | Activo (2021+) | Lightweight, SQL-first, generics |
| **SQLX** | ~16k | Muy maduro (2013+) | SQL helper, no es ORM propiamente |
| **sqlc** | ~13k | Activo (2019+) | SQL-first code generation |
| **Quark** | 0 (no publicado) | Pre-release | Generics-first, security-first |

### 3.2 Matriz Comparativa Detallada

| Feature | Quark | GORM | Ent | Bun | SQLBoiler | sqlc |
|---------|-------|------|-----|-----|-----------|------|
| **Type-Safety (Generics)** | ✅ Nativo | ❌ interface{} | ✅ Code-gen | ✅ Generics | ✅ Code-gen | ✅ Code-gen |
| **SQL Injection Guard** | ✅ SQLGuard | ❌ Sin guard | ✅ Implicit | ❌ Manual | ✅ Implicit | ✅ Implicit |
| **Dialectos** | 6 (PG,MySQL,SQLite,MSSQL,Oracle,MariaDB) | 5 (sin Oracle nativo) | 4 (PG,MySQL,SQLite,Gremlin) | 3 (PG,MySQL,SQLite) | 4 (PG,MySQL,SQLite,MSSQL) | 2 (PG,MySQL) |
| **Oracle Support** | ✅ | ❌ (community) | ❌ | ❌ | ❌ | ❌ |
| **MariaDB (dedicado)** | ✅ Dialecto propio | ❌ Usa MySQL | ❌ | ❌ | ❌ | ❌ |
| **MSSQL** | ✅ | ✅ | ❌ | ❌ | ✅ | ❌ |
| **Auto-Migration** | ✅ | ✅ | ✅ (mejor) | ✅ | ❌ | ❌ |
| **Evolutionary Sync** | ✅ (Add/Rename/Drop) | ✅ (solo add) | ✅ (hash-based diff) | ❌ | ❌ | ❌ |
| **Versioned Migrations** | ⚠️ (básico) | ❌ (externo) | ✅ (Atlas) | ✅ (CLI) | ❌ (externo) | ❌ (externo) |
| **Eager Loading** | ✅ Preload | ✅ Preload + Joins | ✅ With() | ✅ Relations | ✅ Load | ❌ |
| **N+1 Prevention** | ✅ IN query | ✅ (parcial) | ✅ (graph traversal) | ✅ | ✅ | ✅ (compile-time) |
| **Soft Delete** | ✅ Automático | ✅ Plugin | ✅ Policy | ❌ Manual | ❌ Manual | ❌ |
| **Hooks** | ✅ Interface-based | ✅ (más hooks) | ✅ (interceptors) | ✅ | ✅ | ❌ |
| **Validation** | ✅ validator/v10 | ❌ Externo | ✅ Built-in | ❌ Externo | ❌ Externo | ❌ |
| **Multi-Tenant** | ✅ 3 estrategias | ❌ (manual) | ❌ (manual) | ❌ | ❌ | ❌ |
| **Transactions** | ✅ (callback + manual) | ✅ | ✅ | ✅ | ✅ | ✅ |
| **Savepoints** | ✅ | ✅ | ✅ | ✅ | ❌ | ❌ |
| **Nested Tx** | ✅ | ✅ | ✅ | ❌ | ❌ | ❌ |
| **Middleware** | ✅ Pipeline | ✅ Callbacks | ✅ Interceptors | ✅ Hooks | ❌ | ❌ |
| **OpenTelemetry** | ✅ Built-in | ✅ Plugin | ❌ | ✅ Plugin | ❌ | ❌ |
| **Cache L2** | ✅ (Memory + Redis) | ❌ (externo) | ❌ | ❌ | ❌ | ❌ |
| **JSON Queries** | ✅ WhereJSON | ✅ (raw) | ✅ | ✅ | ❌ | ✅ |
| **Composite PK** | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| **Procedures/Functions** | ✅ Routine[T] + Call | ❌ (raw only) | ❌ | ❌ (raw) | ❌ | ✅ |
| **CLI** | ✅ (model gen, inspect, migrate) | ❌ | ✅ (Atlas CLI) | ✅ (migrate) | ✅ (boil) | ✅ (sqlc) |
| **Immutable Builder** | ✅ Clone pattern | ❌ Mutable | ✅ | ✅ | N/A | N/A |
| **Connection Pooling** | ✅ (via sql.DB) | ✅ | ✅ | ✅ | ✅ | ✅ |
| **Read/Write Split** | ❌ | ✅ DBResolver | ❌ | ❌ | ❌ | ❌ |
| **Upsert** | ❌ | ✅ | ✅ | ✅ | ✅ | ✅ |
| **Batch Insert** | ❌ | ✅ | ✅ | ✅ | ✅ | ❌ |
| **Raw SQL** | ✅ (con guard) | ✅ | ✅ | ✅ | ✅ | ✅ |
| **Context Support** | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| **DB Events (PubSub)** | ⚠️ Stub | ❌ | ❌ | ❌ | ❌ | ❌ |
| **Schema Graph Viz** | ❌ | ❌ | ✅ | ❌ | ❌ | ❌ |
| **Dependency** | Ligero | Pesado | Pesado (code-gen) | Ligero | Code-gen | Code-gen |

### 3.3 Donde QUARK VENCE 🏆

1. **Seguridad por Diseño**: SQLGuard es único. Ningún otro ORM en Go tiene validación de identificadores y operadores integrada a nivel del builder. GORM permite inyección SQL trivial con `.Where("name = ? OR 1=1", ...)`.

2. **Multi-Tenant Nativo**: Tres estrategias de aislamiento (DB-per-tenant con LRU, Schema, RLS) como primera clase. Ningún competidor ofrece esto.

3. **Amplitud de Dialectos**: 6 dialectos incluyendo Oracle y MariaDB (con features dedicadas). Solo GORM se acerca pero sin Oracle nativo ni MariaDB dedicado.

4. **Cache L2 Integrado**: Caché semántica con invalidación por tags. Ningún competidor tiene esto built-in.

5. **Generics SIN Code-Gen**: API type-safe con generics nativos de Go, sin paso de generación de código. Ent, SQLBoiler y sqlc requieren code-gen.

6. **Procedures/Functions**: Soporte de primera clase para Routine[T] y Call con OUT params. Prácticamente ningún competidor lo tiene.

7. **Immutable Query Builder**: Clone pattern correcto. GORM tiene problemas conocidos de state mutation.

8. **Validación Integrada**: Dual (tags + interface), interceptada automáticamente en Create. Único en el ecosistema.

9. **MariaDB Features**: Sequences, temporal tables, JSON_TABLE, RETURNING — features específicas que nadie más ofrece.

### 3.4 Donde QUARK PIERDE 🔴

1. **Upsert (INSERT ON CONFLICT)**: **No implementado.** Todos los competidores relevantes lo soportan. Es una operación crítica para producción.

2. **Batch Insert**: No existe `CreateBatch([]T)`. Cada insert es individual. GORM, Bun, Ent todos soportan batch inserts eficientes.

3. **Read/Write Splitting**: No existe routing automático a read-replicas. GORM lo tiene via DBResolver.

4. **Subqueries**: No hay soporte para subqueries en WHERE/FROM/JOIN. GORM y Bun lo soportan.

5. **Group By / Having / Aggregations**: No hay `GroupBy()`, `Having()`, `Sum()`, `Avg()`, `Max()`, `Min()`. GORM, Bun y sqlc lo soportan.

6. **Scopes / Named Queries**: No hay equivalente a los Scopes de GORM para queries reutilizables.

7. **Association Management**: No hay `AddAssociation`, `RemoveAssociation`, `ReplaceAssociations`. GORM tiene gestión completa de asociaciones.

8. **Database Views**: No hay soporte para views.

9. **Index Management**: No hay `CreateIndex`, `DropIndex` en migraciones.

10. **Foreign Key Constraints**: Las migraciones no generan `FOREIGN KEY` constraints.

11. **NOT NULL / DEFAULT**: Las migraciones no respetan tags de nullable/default. Ent genera schemas mucho más ricos.

12. **Optimistic Locking**: No hay soporte para versionado de filas (campo `version`). GORM lo tiene como plugin.

13. **Bulk Update/Delete**: `UpdateMap` requiere WHERE. No hay `UpdateAll`. `DeleteBy` requiere WHERE. No hay forma de hacer bulk operations sin condiciones (safety feature, pero limita uso legítimo).

14. **Query Logging Formatado**: El logger integrado usa `slog` pero no hay un SQL pretty-printer ni slow-query detection automática.

15. **Madurez del Ecosistema**: 0 stars, 0 users, 0 issues, 0 documentación externa, 0 ejemplos en la comunidad.

16. **Documentación en Inglés**: README principal en español. Para alcance global necesita documentación en inglés.

---

## PARTE 4: GAPS CRÍTICOS PARA PUBLICACIÓN

### 4.1 Bloqueadores (Deben resolverse antes de publicar)

| # | Gap | Prioridad | Esfuerzo |
|---|-----|-----------|----------|
| 1 | **Upsert (INSERT ON CONFLICT/MERGE)** | P0 | 2-3 días |
| 2 | **Batch Insert** (`CreateBatch([]T)`) | P0 | 1-2 días |
| 3 | **Corregir bug Paginate** (muta sin clonar) | P0 | 30 min |
| 4 | **Corregir bug Redis doble prefijo** cache key | P0 | 30 min |
| 5 | **Corregir README**: Update() es partial, no full | P0 | 30 min |
| 6 | **Migrator versioned: placeholders dialect-aware** | P0 | 1 día |
| 7 | **Introspection: añadir MariaDB** handler | P0 | 1 hora |
| 8 | **Tests SQLGuard unitarios** | P0 | 1 día |
| 9 | **Tests Migrator versioned** (Up/Down/DryRun) | P0 | 1 día |
| 10 | **Fix loadRelations race** (AllowRawQueries toggle) | P0 | 1 hora |
| 11 | **Fix TestMiddlewareChain** (test roto) | P0 | 30 min |
| 12 | **Fix memory cache goroutine leak** (añadir Close()) | P0 | 30 min |
| 13 | **Eliminar EventBus stub** o marcarlo explícitamente experimental | P0 | 30 min |
| 14 | **Limits enforcement**: MaxWhereConditions y MaxQueryLength no se verifican | P0 | 2 horas |
| 15 | **Errors unused**: ErrTimeout, ErrConstraintViolation definidos pero nunca usados | P0 | 2 horas |

### 4.2 Importantes (Deberían resolverse para V1)

| # | Gap | Prioridad | Esfuerzo |
|---|-----|-----------|----------|
| 16 | **Group By / Having** | P1 | 1-2 días |
| 17 | **Aggregate functions** (Sum, Avg, Max, Min) | P1 | 1 día |
| 18 | **Subqueries** | P1 | 2-3 días |
| 19 | **Index creation en migraciones** | P1 | 1-2 días |
| 20 | **FK constraints en migraciones** | P1 | 1-2 días |
| 21 | **NOT NULL / DEFAULT tags** en schema | P1 | 2 días |
| 22 | **Test Polymorphic E2E** | P1 | 1 día |
| 23 | **Test M2M Preload** | P1 | 1 día |
| 24 | **Test RightJoin** | P1 | 30 min |
| 25 | **Documentación en inglés** | P1 | 2-3 días |
| 26 | **godoc completo** en todas las funciones públicas | P1 | 2 días |
| 27 | **Scopes** (queries reutilizables / named queries) | P1 | 1 día |
| 28 | **WhereNot** | P1 | 2 horas |
| 29 | **Distinct** | P1 | 2 horas |

### 4.3 Nice-to-Have (Post V1)

| # | Gap | Prioridad | Esfuerzo |
|---|-----|-----------|----------|
| 30 | **Read/Write Split** | P2 | 3-5 días |
| 31 | **Optimistic Locking** | P2 | 2 días |
| 32 | **Association Management** (Add/Remove/Replace) | P2 | 2-3 días |
| 33 | **Schema Graph Visualization** | P2 | 3 días |
| 34 | **Slow Query Detection** | P2 | 1 día |
| 35 | **Connection Health Check** | P2 | 1 día |
| 36 | **Database Views support** | P2 | 2 días |
| 37 | **Full-text search helpers** | P2 | 2-3 días |
| 38 | **singleflight en TenantRouter** | P2 | 2 horas |
| 39 | **Real EventBus** (PostgreSQL LISTEN/NOTIFY) | P2 | 3-5 días |
| 40 | **Pluck** (extract single column as slice) | P2 | 2 horas |

---

## PARTE 5: RESUMEN EJECUTIVO

### Fortalezas Diferenciadoras de Quark

Quark tiene un **posicionamiento único** en el ecosistema Go ORM:

1. Es el **único ORM con generics nativos** que no requiere code-gen y cubre 6 dialectos.
2. La **seguridad por diseño** (SQLGuard) es una ventaja real que ningún competidor ofrece.
3. **Multi-tenant nativo** con 3 estrategias es un diferenciador de mercado significativo.
4. **Cache L2 integrado** con invalidación semántica es innovador.
5. El soporte de **Oracle + MariaDB dedicado** abre un mercado enterprise donde GORM no llega.

### Estado Real para Publicación

**No está listo para publicar como "production-ready V1.0".** Las razones principales:

- 7 bugs confirmados que necesitan fix inmediato
- ~30-35% de funcionalidad sin tests adecuados
- Features críticas ausentes (Upsert, Batch Insert, GroupBy)
- Discrepancias documentación vs realidad
- README en español para audiencia global

### Plan de Acción Recomendado

**Sprint 1 (1 semana)**: Fix bugs + tests críticos (items P0: 1-15)  
**Sprint 2 (1 semana)**: Features faltantes clave (Upsert, Batch, GroupBy, Aggregates)  
**Sprint 3 (1 semana)**: Tests + Documentación inglés + godoc  
**Sprint 4 (3 días)**: Release prep, examples, CI pipeline

**Fecha estimada de publicación**: 4 semanas desde hoy.

---

*Documento generado tras auditoría completa de 50+ archivos fuente, ~8000 líneas de código de producción, ~4500 líneas de tests, y toda la documentación disponible.*
