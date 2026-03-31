# Release Checklist

Fecha de referencia: 2026-03-31.

## 0) Toolchain

- Minimo contractual de compilacion: `go 1.23` (`go.mod`).
- Toolchain recomendada para release: `Go 1.26.x`.
- GoReleaser usado en pipeline: `v2.14.1`.
- Politica completa: `docs/GO_VERSION_POLICY.md`.

## 1) Gate de calidad (obligatorio)

Ejecutar antes de etiquetar release:

```bash
./scripts/release/rehearse_rc.sh
```

Este script ejecuta:

1. `go test ./...`
2. `go test ./examples/mvc_api -run TestExampleMVCAPIAdmin_Smoke -v`
3. `node --check pkg/admin/ui/components.js`
4. `node --check pkg/admin/ui/app.js`
5. `goreleaser check`
6. `goreleaser release --snapshot --clean --skip=publish --skip=announce`

## 2) Rehearsal en GitHub Actions (opcional recomendado)

Lanzar manualmente:

- Workflow: `.github/workflows/rehearsal.yml`
- Trigger: `workflow_dispatch`

Objetivo: repetir el rehearsal en runner limpio sin publicar artefactos.

## 3) Documentacion (obligatorio)

Comprobar coherencia entre codigo y docs:

- `README.md`
- `CHANGELOG.md`
- `docs/VERSIONING.md`
- `docs/QUICKSTART.md`
- `docs/TUTORIAL_DETALLADO.md`

## 4) Tagging y publicacion

1. Actualizar `CHANGELOG.md` (mover `Unreleased` al nuevo tag).
2. Crear tag SemVer:
   - estable: `v0.x.y`
   - pre-release: `v0.x.y-rcN`
3. Push del tag para disparar `.github/workflows/release.yml`.

## 5) Verificacion post-release

1. Confirmar assets para:
   - `linux/amd64`, `linux/arm64`
   - `darwin/amd64`, `darwin/arm64`
   - `windows/amd64`, `windows/arm64`
2. Confirmar presencia de `checksums.txt`.
3. Validar binario:

```bash
./goframe version
```

Debe imprimir version de release (no `dev`).
