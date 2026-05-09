# GoFrame

[![CI](https://github.com/jcsvwinston/nucleus/actions/workflows/ci.yml/badge.svg)](https://github.com/jcsvwinston/nucleus/actions/workflows/ci.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/jcsvwinston/nucleus.svg)](https://pkg.go.dev/github.com/jcsvwinston/nucleus)

**Enterprise web framework for Go** — Simple like Gin, powerful like Django.

## Why GoFrame

- **MVC + REST API**: Build web apps and APIs with the same framework
- **Auto-generated Admin**: CRUD interface from your models
- **SQL-first**: Direct `database/sql` with migrations, no ORM magic
- **Production-ready**: Multi-tenancy, background jobs, observability built-in
- **Go-idiomatic**: stdlib-first design with minimal dependencies

## Quick Start

### Install CLI

```bash
go install github.com/jcsvwinston/nucleus/cmd/nucleus@latest
```

### Create Project

```bash
goframe new myapp
cd myapp
go mod tidy
go run ./cmd/server
```

Open:
- `http://localhost:8080/` — Web app
- `http://localhost:8080/api/articles` — API
- `http://localhost:8080/admin` — Admin panel

## Simple API Example

```go
package main

import (
    "github.com/jcsvwinston/nucleus/pkg/goframe"
)

type Article struct {
    ID    int64  `json:"id" db:"id"`
    Title string `json:"title" db:"title" validate:"required"`
}

func main() {
    goframe.New().
        Port(8080).
        SQLite("app.db").
        Model(&Article{}).
        AutoMigrate().
        Get("/api/articles", func(c *goframe.Context) error {
            return c.JSON(200, []Article{{ID: 1, Title: "Hello"}})
        }).
        Run()
}
```

## Project Structure

Generated projects follow standard layout:

```
myapp/
├── cmd/server/main.go      # HTTP server
├── internal/
│   ├── models/             # Domain models
│   ├── controllers/        # HTTP handlers
│   ├── services/           # Business logic
│   └── repositories/       # Data access
├── migrations/             # SQL migrations
└── nucleus.yml           # Configuration
```

## Documentation

- **[docs/README.md](docs/README.md)** — Documentation index
- **[docs/QUICKSTART.md](docs/QUICKSTART.md)** — 5-minute quickstart
- **[docs/guides/DETAILED_TUTORIAL.md](docs/guides/DETAILED_TUTORIAL.md)** — Complete tutorial
- **[docs/reference/PROJECT_LAYOUT.md](docs/reference/PROJECT_LAYOUT.md)** — Project structure

## Key Features

| Feature | Description |
|---------|-------------|
| **Models** | Struct-based with validation, hooks, admin metadata |
| **Migrations** | SQL files with up/down, CLI-managed |
| **Admin Panel** | Auto-generated from registered models |
| **Auth** | JWT, sessions (memory/SQL/Redis), Casbin RBAC |
| **Tasks** | Background jobs with Asynq + Redis |
| **Outbox** | Transactional outbox pattern |
| **Multi-tenancy** | Subdomain/header based with DB isolation |
| **Storage** | S3, GCS, Azure, local abstractions |
| **Observability** | OpenTelemetry, structured logging |

## CLI Commands

```bash
goframe new myapp          # Create project
goframe serve              # Run server
goframe migrate            # Run migrations
goframe seed               # Load seed data
goframe createuser         # Create admin user
goframe routes             # List routes
goframe health             # Check health
```

## Requirements

- Go 1.25+
- SQLite/PostgreSQL/MySQL
- Optional: Redis (for tasks/sessions)

## License

MIT — See [LICENSE](LICENSE)
