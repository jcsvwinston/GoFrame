# Go Version Policy

Fecha de referencia: 2026-03-31.

## Objetivo

Equilibrar estabilidad para usuarios del framework y adopcion progresiva de nuevas versiones de Go.

## Niveles de soporte

1. Minimo soportado (contrato de compilacion): `go 1.23` (definido en `go.mod`).
2. Recomendado para desarrollo local: `Go 1.26.x`.
3. Toolchain de release CI: `Go 1.26.x`.

## CI y compatibilidad

- CI ejecuta pruebas de Go en matriz `1.23.x` y `1.26.x`.
- Smoke test y checks de UI se ejecutan en `1.26.x`.
- La release por tag usa `1.26.x` y GoReleaser `v2.14.1`.

## Plan de elevacion del minimo

Para subir el minimo de `go.mod` a una version superior:

1. Mantener durante al menos 2 ciclos de release la matriz con la nueva version.
2. Verificar quickstart y ejemplo oficial en la nueva toolchain.
3. Anunciarlo en `CHANGELOG.md` y docs de fase antes del cambio.

## Decision para Fase 5

- No se eleva aun el minimo contractual (`go.mod`).
- Si no aparecen incidencias en siguientes ciclos, objetivo: considerar subida del minimo en `v0.6.0`.
