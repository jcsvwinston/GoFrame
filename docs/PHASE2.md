# Fase 2 - Migracion SQL a Bun (Cerrada)

## Objetivo

Consolidar Bun como runtime SQL principal del framework manteniendo una via de compatibilidad temporal para codigo legado en GORM.

## Estrategia

Migracion por slices pequenos con compatibilidad temporal:

1. Introducir un seam de engine dual en `pkg/db`.
2. Habilitar seleccion por configuracion (`database_engine`).
3. Migrar componentes internos por modulo (`app` -> `model` -> `admin`).
4. Retirar camino GORM cuando Bun cubra el 100% del contrato requerido.

## Slice 1 (Completado)

- `pkg/db` soporta dos engines:
  - `gorm` (compatibilidad)
  - `bun` (ruta objetivo)
- Nuevo `db.Config.Engine`.
- Nuevos helpers:
  - `Engine()`
  - `BunDB()`
  - `TxBun(...)`
- Errores tipados cruzados:
  - `ErrGORMRequired`
  - `ErrBunRequired`
- Tests de engine dual añadidos.
- `pkg/app.Config` incorpora `DatabaseEngine` (default actual: `bun`).
- `app.New(...)` propaga engine al constructor de DB.
- `app.New(...)` propaga el engine seleccionado al panel admin.

## Slices

### Slice 2

- Estado: completado.
- `pkg/model` ya incorpora `CRUDBun` paralelo al CRUD actual.
- Tests de integracion iniciales añadidos para create/find/paginacion/update/delete sobre engine Bun.
- `pkg/admin` ya conecta path Bun sin dependencia directa de `*gorm.DB`.
- `app.New(...)` ya inicializa admin tambien en engine Bun.

### Slice 3

- Estado: completado.
- Handlers de `pkg/admin` migrados a capa de acceso desacoplada del ORM via `model.CRUDOperator`.
- Conteos y create/list/update/delete ya funcionan en `gorm` y `bun`.

### Slice 4

- Estado: completado.
- `bun` activado como default en `app.Config.defaults()`.
- `gorm` mantenido en modo compatibilidad temporal.

### Slice 5

- Estado: completado.
- Hooks de `model.ModelConfig` desacoplados de `*gorm.DB` mediante `HookContext` agnostico de engine.
- `db.AutoMigrate(...)` protegido para retornar error tipado en runtime Bun.
- `model.BaseModel` deja de depender de `gorm.DeletedAt`.
- `ExtractMeta` soporta tag neutral `db:` manteniendo compatibilidad con `gorm:`.
- `pkg/db/migrate` incorpora operaciones reales `Up()`, `Down()`, `Steps(n)` y `Status()` sobre archivos `.up.sql` / `.down.sql`.
- `cmd/goframe` + `internal/cli` entregan baseline de CLI tipo `manage.py` con:
  - `serve`, `migrate`, `createuser`, `seed`, `shell`, `generate`, `routes`, `health`.
  - `migrate` con `up|down|steps|status|create|reset|refresh`.
  - guardrails de produccion en acciones destructivas (`--force` / `--yes`).
  - extensibilidad por comandos externos `goframe-<nombre>` en `PATH`.
  - `generate resource` para scaffold CRUD base (modelo + handler + test + migracion).
- `db.New(...)` utiliza Bun por defecto.
- Fallback implicito a GORM eliminado en rutas internas de `admin` (switch de engine explicito).
- Documentacion y quickstarts alineados a Bun-first.

## Estado Final

- Fase 2 cerrada a fecha 2026-03-23.
- Runtime principal validado en Bun (`go test ./...` en verde).
- `gorm` queda en modo compatibilidad temporal y explicita, fuera del path principal recomendado.

## Criterio de Cierre de Fase 2

1. [x] Bootstrap por defecto en Bun.
2. [x] `model` y `admin` funcionando sobre Bun.
3. [x] `go test ./...` verde en Bun-first.
