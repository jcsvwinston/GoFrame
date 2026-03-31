# Example: MVC + API + Admin

Este ejemplo abre Fase 4 con una app ejecutable que mezcla:

- pagina MVC (`/`)
- API REST (`/api/articles`)
- panel admin (`/admin`)

## Ejecutar

```bash
go run ./examples/mvc_api
```

Abrir:

- `http://localhost:8090/`
- `http://localhost:8090/api/articles`
- `http://localhost:8090/admin`

## Crear un articulo por API

```bash
curl -X POST http://localhost:8090/api/articles \
  -H "Content-Type: application/json" \
  -d '{"title":"Nuevo","content":"Texto desde API","published":true}'
```
