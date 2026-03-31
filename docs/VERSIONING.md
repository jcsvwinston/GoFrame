# Versioning Strategy

Fecha de referencia: 2026-03-31.

## Objetivo

Definir una estrategia simple y predecible para etiquetar releases de GoFrame hasta `v1.0.0`.

## Regla principal

Se usa SemVer en modo pre-1.0:

- `v0.x.y`
- `x`: incremento por entregas funcionales relevantes o cambios de contrato.
- `y`: fixes y mejoras compatibles sin ruptura intencional.

## Politica para cambios de contrato

Mientras estemos en `v0.x.y`, los cambios incompatibles se permiten, pero deben:

1. mencionarse en `CHANGELOG.md` en seccion `Changed` o `Removed`,
2. reflejarse en docs de fase/quickstart,
3. subir al menos el minor (`x`).

## Cadencia sugerida

- Release tecnica cada cierre de fase relevante.
- Patch release para fixes criticos entre fases.

## Tagging

- Formato de tag: `v0.x.y`
- Formato pre-release: `v0.x.y-rcN`
- Ejemplos:
  - `v0.4.0` cierre Fase 4
  - `v0.4.1` fix compatible post-release
  - `v0.5.0-rc1` primer candidato de release
  - `v0.5.0` release estable tras rehearsal

## Criterio de paso a v1.0.0

1. Contratos estables en `pkg/app`, `pkg/model`, `pkg/admin`.
2. CLI base estable (`new`, `serve`, `migrate`, `seed`, `createuser`, `routes`, `health`).
3. Quickstart/tutorial y ejemplo oficial mantenidos y validados por CI.
