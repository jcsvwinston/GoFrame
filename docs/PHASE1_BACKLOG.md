# Fase 1 - Backlog Tecnico Ejecutable

Objetivo de Fase 1: construir el nucleo de aplicacion (`pkg/app/app.go`) con lifecycle claro y wiring minimo estable.

## Estado (2026-03-23)

- Completado: A1, A2, A3, B1, B2, B3, B4, C1, C2, C3.
- Fase 2 abierta: plan de migracion tecnica en [docs/PHASE2.md](/Users/jcsv/GolandProjects/GoFrame/GoFrame/docs/PHASE2.md).

## Epic A: App Container

### A1. Crear `pkg/app/app.go` con struct `App`

- Incluye: `Config`, `Logger`, `Router`, `DB`, `Models`, `Admin`.
- Criterio de aceptacion: existe constructor `New(cfg *Config) (*App, error)` con validaciones basicas.

### A2. Implementar `OnShutdown` y pipeline de cierre ordenado

- Registrar funciones de cleanup.
- Ejecutar cierre en orden inverso con propagacion de errores.
- Criterio de aceptacion: tests que verifiquen orden y ejecucion aun con error intermedio.

### A3. Implementar `Run(ctx)` con servidor HTTP y senales

- Arranque de `http.Server`.
- Escucha de SIGINT/SIGTERM.
- Graceful shutdown con timeout configurable.
- Criterio de aceptacion: test de integracion que valida cierre limpio.

## Epic B: Wiring Minimo

### B1. Inicializar logger y router por defecto

- `observe.NewLogger(cfg.LogLevel, cfg.LogFormat)`.
- `router.New(...)` aplicado en `New`.
- Criterio de aceptacion: `App` funcional con configuracion por defecto.

### B2. Inicializar capa SQL en `New`

- Integrar `pkg/db.New`.
- Registrar healthcheck de DB.
- Criterio de aceptacion: falla temprano si `DatabaseURL` invalida.

### B3. Inicializar `model.Registry` y exponer helper de registro

- Metodo sugerido: `RegisterModel(m any, cfg ...model.ModelConfig) error`.
- Criterio de aceptacion: modelos disponibles en `App.Models`.

### B4. Inicializar `admin.Panel` opcional

- Si `cfg.AdminPrefix` esta vacio, no montar admin.
- Si esta definido, crear panel y permitir `MountAdmin()`.
- Criterio de aceptacion: admin montable sin wiring manual repetitivo.

## Epic C: Contrato y DX

### C1. Documentar bootstrap oficial de Fase 1

- Actualizar README con ejemplo `app.New(...).Run(...)`.
- Criterio de aceptacion: ejemplo compilable y alineado con API real.

### C2. Crear test suite de `pkg/app`

- Tests de constructor, run/shutdown, on-shutdown y error paths.
- Criterio de aceptacion: cobertura significativa de rutas criticas.

### C3. Definir errores de inicializacion estandarizados

- Errores envueltos con contexto por componente (`app.New db: ...`).
- Criterio de aceptacion: mensajes accionables para usuario final.

## Orden Recomendado

1. A1
2. B1
3. B2
4. A2
5. A3
6. B3
7. B4
8. C3
9. C2
10. C1
