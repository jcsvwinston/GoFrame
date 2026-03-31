# Fase 3 - Admin UI Rica + DX (En Progreso)

## Objetivo

Llevar el panel admin desde una SPA funcional basica a una experiencia rica y productiva, alineada con la direccion del proyecto: estilo Tailwind, componentes reutilizables y foco en usabilidad real para MVC + API.

## Estado

- Inicio de fase: 2026-03-31.
- Slice 1: completado en este arranque.
- Slice 2: completado (filtros por campo, ordenacion por columnas, export de seleccionados).
- Slice 3: completado (libreria base de componentes y tabs/paneles de detalle).
- Slice 4: completado (accesibilidad de tabs/palette/modales, retries de red y errores recuperables con reintento).
- Scope actual: frontend embebido `pkg/admin/ui` sin romper API backend.

## Slice 1 (Completado)

- Nuevo shell visual del admin:
  - sidebar + topbar + breadcrumbs + vista principal.
- Sistema de componentes UI base:
  - botones, tablas, formularios, chips de estado, modales, toasts.
- Command palette (`Ctrl/Cmd + K`) para navegacion rapida y acciones.
- Estados de carga y vacio consistentes.
- Mejoras responsive para navegacion movil.
- Mantiene el contrato de API actual (`/api/models`, schema, CRUD, bulk, export).

## Slices ejecutados

### Slice 2

- Estado: completado.
- Filtros por campo usando metadata `is_filter` del schema y mapping de query params.
- Ordenacion por columnas con `order_by` saneado en backend.
- Acciones bulk ampliadas con export de seleccionados.

### Slice 3

- Estado: completado.
- Libreria base de componentes frontend (`components.js`) para secciones, estados y tarjetas de detalle.
- Tokens de diseno consolidados en `style.css`.
- Tabs y paneles de detalle en formularios para modelos con mayor densidad de campos.

### Slice 4

- Estado: completado.
- Pulido de accesibilidad:
  - foco visible global, navegacion teclado en tabs y command palette, roles/ARIA refinados en tabs y overlays.
- Hardening de UX:
  - retries para errores transitorios de red/API.
  - estados de error recuperables con accion de reintento.
  - feedback de estado en operaciones largas (`Saving...`, `Deleting...`, `Exporting...`).

## Criterio de cierre de Fase 3

1. UI admin rica y consistente en dashboard/list/form/detail.
2. Navegacion rapida (command palette) y productividad diaria mejorada.
3. Responsive usable en movil y escritorio.
4. Cobertura de tests de comportamiento UI critico (al menos smoke + flujos CRUD principales).
5. Documentacion actualizada (quickstart/tutorial) con flujo de trabajo final de admin.
