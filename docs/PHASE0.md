# Fase 0 - Contrato y Direccion

## Objetivo

Cerrar decisiones de arquitectura y alinear el contrato publico del proyecto antes de expandir funcionalidad.

## Decisiones Cerradas

1. SQL oficial: Bun.
2. MongoDB oficial: `go.mongodb.org/mongo-driver`.
3. Cache y pub/sub: `redis/go-redis/v9`.
4. Admin objetivo: Tailwind CSS + componentes reutilizables.
5. Arquitectura polyglot por interfaces; sin unificar SQL/Mongo/Redis en una API magica.

## Entregables Hechos

1. SPEC actualizado con estrategia polyglot y stack oficial:
   [SPEC.md](/Users/jcsv/GolandProjects/GoFrame/GoFrame/SPEC.md)
2. Estructura objetivo ampliada con `pkg/document`:
   [SPEC.md](/Users/jcsv/GolandProjects/GoFrame/GoFrame/SPEC.md)
3. Config objetivo ampliada con `MongoURL` y `MongoDB`:
   [SPEC.md](/Users/jcsv/GolandProjects/GoFrame/GoFrame/SPEC.md)
4. README alineado al estado real y al target:
   [README.md](/Users/jcsv/GolandProjects/GoFrame/GoFrame/README.md)

## Riesgos Identificados

1. Divergencia temporal entre codigo actual (GORM) y target (Bun).
2. Admin actual sin Tailwind ni set completo de componentes.
3. (Mitigado en iteraciones posteriores) CLI inicialmente incompleta respecto a `manage.py`.

## Criterios de Salida de Fase 0

- [x] Decisiones de stack formalizadas.
- [x] Contrato documental alineado (README + SPEC).
- [x] Diferencia entre estado actual y target explicitada.
- [x] Backlog de Fase 1 convertido en tareas tecnicas ejecutables ([docs/PHASE1_BACKLOG.md](/Users/jcsv/GolandProjects/GoFrame/GoFrame/docs/PHASE1_BACKLOG.md)).

## Entrada a Fase 1

1. Implementar `pkg/app/app.go` como contenedor principal y lifecycle.
2. Definir wiring minimo obligatorio (`Config`, `Logger`, `Router`, `DB`, `Registry`).
3. Establecer plantilla de arranque para proyectos nuevos (base MVC + API).
