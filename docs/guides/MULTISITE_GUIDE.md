# MultiSite & MultiTenant Guide

Reference date: 2026-04-10.
Status: Current.

This guide covers GoFrame's MultiSite and MultiTenant request scope resolution, including site resolution, tenant routing, database alias routing, and security isolation.

## Table of Contents

- [Overview](#overview)
- [MultiSite Configuration](#multisite-configuration)
- [Site Resolution](#site-resolution)
- [MultiTenant Configuration](#multitenant-configuration)
- [Tenant Resolution](#tenant-resolution)
- [Tenant-to-Database Routing](#tenant-to-database-routing)
- [Security Isolation](#security-isolation)
- [Request Context Usage](#request-context-usage)
- [Production Checklist](#production-checklist)

---

## Overview

GoFrame provides request-scope resolution for:

| Feature | Purpose | Resolution Method |
|---------|---------|-------------------|
| **MultiSite** | Multiple sites/domains on one app | Host matching (exact or wildcard) |
| **MultiTenant** | Multiple tenants with data isolation | Subdomain or header-based resolution |

Both features work together and are resolved at the middleware level before reaching your handlers.

---

## MultiSite Configuration

```yaml
# nucleus.yml
multisite:
  enabled: true
  sites:
    - host: "example.com"
      name: "Main Site"
      locale: "en"
    - host: "es.example.com"
      name: "Spanish Site"
      locale: "es"
    - host: "*.example.org"
      name: "Wildcard Site"
      locale: "en"
```

---

## Site Resolution

GoFrame resolves the current site from the request host:

### Exact host matching

```yaml
multisite:
  sites:
    - host: "example.com"
      name: "Main Site"
    - host: "admin.example.com"
      name: "Admin Portal"
```

### Wildcard matching

```yaml
multisite:
  sites:
    - host: "*.example.com"
      name: "Regional Sites"
```

Matches:
- `us.example.com`
- `eu.example.com`
- `asia.example.com`

### Accessing current site in handlers

```go
import "github.com/jcsvwinston/nucleus/pkg/app"

func (h *Handler) Home(w http.ResponseWriter, r *http.Request) {
    scope := app.RequestScopeFromContext(r.Context())
    if scope == nil {
        // No scope resolved (no multisite configured)
        return
    }

    site := scope.Site()
    if site != nil {
        log.Printf("Current site: %s (locale: %s)", site.Name, site.Locale)
    }
}
```

---

## MultiTenant Configuration

```yaml
# nucleus.yml
multitenant:
  enabled: true
  resolver: subdomain       # Options: subdomain, header
  header: X-Tenant-ID       # Only used when resolver=header
  require_isolated_db: true
  tenants:
    acme:
      database: acme_db
    globex:
      database: globex_db
    initech:
      database: initech_db  # Rejected if require_isolated_db=true
```

---

## Tenant Resolution

### Subdomain resolution

The tenant ID is extracted from the subdomain:

```
acme.example.com  -> tenant: "acme"
globex.example.com -> tenant: "globex"
```

```yaml
multitenant:
  resolution: subdomain
```

### Header resolution

The tenant ID is extracted from a request header:

```yaml
multitenant:
  resolution: header
  header_name: X-Tenant-ID
```

```bash
curl -H "X-Tenant-ID: acme" https://api.example.com/articles
```

### Accessing current tenant in handlers

```go
func (h *Handler) GetArticles(w http.ResponseWriter, r *http.Request) {
    scope := app.RequestScopeFromContext(r.Context())
    if scope == nil {
        errors.WriteHTTP(w, errors.NewBadRequest("tenant not resolved"))
        return
    }

    tenant := scope.Tenant()
    if tenant == "" {
        errors.WriteHTTP(w, errors.NewUnauthorized("tenant required"))
        return
    }

    log.Printf("Current tenant: %s", tenant)

    // Query tenant-scoped data
    articles, err := h.repo.FindByTenant(tenant)
    // ...
}
```

---

## Tenant-to-Database Routing

GoFrame maps tenants to database aliases:

### Explicit mapping

```yaml
multitenant:
  tenants:
    acme:
      database: acme_db
    globex:
      database: globex_db
```

### Template-based mapping

For large numbers of tenants, use templates:

```yaml
multitenant:
  database_alias_template: "tenant_%s"
```

This maps:
- `acme` -> `tenant_acme`
- `globex` -> `tenant_globex`

### Using tenant-scoped database in handlers

```go
func (h *Handler) GetArticles(w http.ResponseWriter, r *http.Request) {
    // Get database for current tenant
    db := app.DatabaseForRequest(r)
    if db == nil {
        errors.WriteHTTP(w, errors.NewInternal("tenant database not configured"))
        return
    }

    // Or use explicit alias
    db = app.Database("acme_db")

    // Query with tenant-scoped DB
    rows, err := db.QueryContext(r.Context(), "SELECT * FROM articles WHERE tenant_id = ?", tenantID)
}
```

---

## Security Isolation

### `require_isolated_db` guardrail

When `require_isolated_db: true` (default), GoFrame enforces:

1. **Startup validation**: Rejects configuration if multiple tenants map to the same database alias.
2. **Request routing**: Rejects requests when tenant isolation is required but the resolved database is shared.

```yaml
multitenant:
  require_isolated_db: true  # Recommended for production
```

### What this prevents

```yaml
# INVALID when require_isolated_db=true
multitenant:
  require_isolated_db: true
  tenants:
    acme:
      database: shared_db  # REJECTED: multiple tenants on same DB
    globex:
      database: shared_db  # REJECTED
```

### When to disable isolation

Only disable for specific use cases:

```yaml
multitenant:
  require_isolated_db: false  # Use with caution
  tenants:
    internal:
      database: main_db  # Internal tenant shares main DB
```

---

## Request Context Usage

### Full scope access

```go
func (h *Handler) Handle(w http.ResponseWriter, r *http.Request) {
    scope := app.RequestScopeFromContext(r.Context())
    if scope == nil {
        return // No multisite/multitenant configured
    }

    // Site information
    if scope.Site() != nil {
        fmt.Printf("Site: %s, Locale: %s\n", scope.Site().Name, scope.Site().Locale)
    }

    // Tenant information
    if scope.Tenant() != "" {
        fmt.Printf("Tenant: %s\n", scope.Tenant())
    }

    // Database for tenant
    db := scope.Database()
    if db != nil {
        fmt.Printf("Database alias: %s\n", db.Alias())
    }
}
```

### Helper functions

```go
// Get database for current request's tenant
db := app.DatabaseForRequest(r)

// Get database by explicit alias
db := app.Database("analytics")

// Get primary database
db := app.DB
```

---

## Production Checklist

- [ ] `multisite.enabled: true` configured with explicit site definitions
- [ ] Wildcard hosts use valid patterns (`*.example.com`)
- [ ] `multitenant.enabled: true` when tenant isolation is required
- [ ] `multitenant.require_isolated_db: true` (production default)
- [ ] All tenants have explicit database mappings or template configured
- [ ] Tenant resolution method chosen (`subdomain` or `header`)
- [ ] Handlers check for nil scope before accessing tenant data
- [ ] Database queries use tenant-scoped connections
- [ ] Admin panel secured with tenant-aware session store
- [ ] Health checks validate tenant database connectivity
