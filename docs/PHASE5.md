# Fase 5 - Release Candidate (Cerrada)

## Objetivo

Dejar operativa la cadena completa de release: validacion, versionado y publicacion automatizada de artefactos multiplataforma.

## Estado

- Inicio de fase: 2026-03-31.
- Cierre de fase: 2026-03-31.
- Resultado: completada.

## Slice 1 (Completado)

- `CHANGELOG.md` inicial con `Unreleased` y `0.4.0`.
- CI en `.github/workflows/ci.yml`.
- Release por tags en `.github/workflows/release.yml`.
- Estrategia de versionado en `docs/VERSIONING.md`.
- Checklist operativo en `docs/RELEASE_CHECKLIST.md`.

## Slice 2 (Completado)

- Publicacion automatizada de binarios con GoReleaser:
  - configuracion en `.goreleaser.yaml`
  - builds para `linux/darwin/windows` y `amd64/arm64`
  - checksums SHA256 en `checksums.txt`
- Workflow de release actualizado para publicar artefactos con `goreleaser/goreleaser-action@v6`.

## Slice 3 (Completado)

- Rehearsal end-to-end para RC:
  - script `scripts/release/rehearse_rc.sh`
  - workflow manual `.github/workflows/rehearsal.yml` (`workflow_dispatch`)
- Validacion local ejecutada con exito:
  1. `go test ./...`
  2. smoke test del ejemplo MVC/API/Admin
  3. checks de sintaxis JS del admin
  4. `goreleaser check`
  5. `goreleaser release --snapshot --clean --skip=publish --skip=announce`

## Criterio de cierre de Fase 5

1. CI obligatorio para PRs y rama principal: cumplido.
2. Release automatizada por tag SemVer (`v0.x.y` y pre-releases `v0.x.y-rcN`): cumplido.
3. Changelog y versionado activos en flujo diario: cumplido.
4. Ensayo de release E2E para `v0.5.0-rc1`: cumplido (modo snapshot).
