# Fase 4 - Ejemplos MVC/API + Release (Cerrada)

## Objetivo

Cerrar la brecha final de adopcion: que un desarrollador pueda arrancar una app real viendo codigo ejecutable dentro del repo, no solo documentacion.

## Estado

- Inicio de fase: 2026-03-31.
- Slice 1: completado.
- Slice 2: completado (starter CLI `goframe new`).
- Slice 3: completado (smoke E2E + checklist de release/versionado).
- Estado final: fase cerrada.

## Slice 1 (Completado)

- Nuevo ejemplo ejecutable en `examples/mvc_api`:
  - pagina MVC (`/`)
  - API REST (`/api/articles`)
  - admin embebido (`/admin`)
- Misma entidad (`Article`) expuesta en MVC, API y Admin para demostrar integracion real.
- Seed inicial y schema autocreado para onboarding rapido.

## Slices ejecutados

### Slice 2

- Estado: completado.
- Plantilla de proyecto inicial oficial (`starter`) orientada a equipos.
- Nuevo comando `goframe new` para bootstrap de app completa MVC + API + Admin.

### Slice 3

- Estado: completado.
- Endurecimiento de release:
  - smoke tests E2E del ejemplo en `examples/mvc_api/main_test.go`.
  - checklist de release y versionado inicial en `docs/RELEASE_CHECKLIST.md`.

## Criterio de cierre de Fase 4

1. Al menos un ejemplo oficial runnable MVC + API + Admin.
2. Ruta de arranque clara para nuevos proyectos.
3. Documentacion sincronizada con codigo real del repo.
