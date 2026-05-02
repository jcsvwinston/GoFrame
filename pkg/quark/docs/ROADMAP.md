# Quark ORM Roadmap

This document outlines the current state and future development goals for the Quark ORM.

## Completed Features (v0.1)

- [x] **Type-Safe API**: Generic-based `Query[T]` builders.
- [x] **Database Dialects**: Initial support for SQLite, PostgreSQL, and MySQL.
- [x] **Nested Transactions**: Support for transactions and Savepoints.
- [x] **Eager Loading**: Prevent N+1 queries using `.Preload()`.
- [x] **Lifecycle Hooks**: Interfaces for `BeforeCreate`, `AfterUpdate`, etc.
- [x] **Model Validation**: Tag-based and programmatic struct validation via `validator/v10`.
- [x] **Schema Migrations**: Automatic table creation based on struct fields via `client.Migrate()`.
- [x] **Multi-Tenant Routing**: `TenantRouter` supporting Database-per-tenant, Schema-per-tenant, and Row-level strategies.
- [x] **Data Streaming**: Efficient iteration over large datasets via `.Iter()`.

## Completed Features (v0.2)

- [x] **Quark CLI**: Standalone tool for project initialization and code generation.
- [x] **Advanced Migrations**: Version-controlled migration files (Up/Down) via CLI.
- [x] **Model Introspection**: Automated generation of Go models from existing DB schemas.
- [x] **Many-to-Many Relations**: Support for M2M associations via join tables with `m2m` tags.
- [x] **Polymorphic Relations**: Support for polymorphic associations with `polymorphic` tags.
- [x] **Custom Dialects**: Public API to register custom database dialects via `RegisterDialect()`.
- [x] **Evolutionary Migrations**: ALTER TABLE support for adding, dropping, and modifying columns.

## In Progress / Short-Term Goals

- [ ] **Observability**: Add native OpenTelemetry tracing (Spans) and metrics.
- [ ] **Level 2 Query Caching**: Integrate a fast in-memory or Redis-backed cache layer for complex, read-heavy query patterns.

- [ ] **Extended Dialects**: Enhance Dialect implementations to natively support Microsoft SQL Server and Oracle databases.
- [ ] **JSON Fields**: Add robust support for querying and mutating JSON/JSONB fields in structs.

## Long-Term Goals

- [ ] **Standalone GoFrame Module**: Release Quark as an entirely decoupled `go-quark` module outside the GoFrame core.
- [ ] **Read/Write Splitting**: Automatic routing to read-replicas for SELECT queries.
- [ ] **Query Optimizer Hints**: Add specific builder methods to force index usage.
