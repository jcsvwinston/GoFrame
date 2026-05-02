# Quark ORM v1.0.0 Release Notes

We are proud to announce the release of **Quark ORM v1.0.0**, a production-ready, high-performance ORM for Go.

## Key Features in v1.0.0

### 1. Robust Multi-Tenant Isolation
*   **Automatic Tenant Injection**: Support for `RowLevelSecurity` (RLS) strategies that automatically inject tenant identifiers into queries and mutations.
*   **Recursive Isolation**: Tenant context is correctly propagated through complex recursive saving of nested associations (has_one, has_many, belongs_to).
*   **Database-per-Tenant**: Support for routing queries to different database schemas or instances based on tenant context.

### 2. Advanced Persistence Engine
*   **Recursive Association Saving**: Deeply nested entity graphs can be saved in a single call, with intelligent handling of mixed new (INSERT) and existing (UPDATE) records.
*   **Smart Join Inference**: Automatic detection of foreign key columns based on naming conventions (e.g., `Author` -> `author_id`), with support for explicit overrides.
*   **Atomic Transactions**: Full support for transactions via the `BeginTx` API, ensuring data integrity during complex multi-step operations.

### 3. Production-Grade Observability
*   **Centralized Query Observers**: Comprehensive monitoring of all database interactions (Select, Insert, Update, Delete) with detailed timing and metadata.
*   **SQL Query Logging**: Built-in structured logging for slow queries and execution errors.
*   **Middleware Pipeline**: Flexible middleware chain for intercepting and augmenting query execution (caching, auditing, retries).

### 4. Enterprise Database Support
*   **Multi-Dialect Compatibility**: First-class support for **SQLite**, **PostgreSQL**, **MySQL**, **MSSQL**, and **Oracle**.
*   **Safe Auto-Migrations**: Intelligent schema synchronization that detects column changes, renames, and type mismatches while protecting against accidental data loss.

## Performance Benchmarks
Quark ORM v1.0.0 has been optimized for low-latency reflection and efficient SQL generation, achieving performance parity with raw SQL in many scenarios through aggressive metadata caching.

## Getting Started
Refer to the `examples/` directory for dialect-specific samples and `docs/ARCHITECTURE.md` for a deep dive into the internal design.

---
*Quark ORM - Small, Fast, and Powerful.*
