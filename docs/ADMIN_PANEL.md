# Admin Panel Documentation

Reference date: 2026-04-11.
Status: Current (v0.7.x).

This document describes the GoFrame admin panel capabilities, configuration, and usage.

## Overview

The GoFrame admin panel (`/admin`) is an embedded SPA interface that provides:

- Data management (CRUD) for registered models
- Real-time runtime inspection (traffic, SQL, sessions)
- System health monitoring
- Multi-tenant and multi-site management
- Role-based access control (RBAC)
- Audit logging
- Background job monitoring
- Migration management
- Deployment awareness (standalone, Docker, K8s, clusters)

## Configuration

### Basic Configuration

```yaml
admin_prefix: /admin           # URL prefix (default: /admin)
admin_title: GoFrame Admin     # Panel title
admin_auth_database: default   # Database for admin users table
admin_bootstrap_username: admin
admin_bootstrap_email: admin@localhost
admin_bootstrap_password: ""   # Empty = auto-generate on first run
```

### Multi-Tenant Configuration

When `multitenant.enabled: true`, the admin panel automatically:

1. Detects tenant fields in models (`db:"tenant"` tag or `tenant_id` column convention)
2. Filters CRUD queries by the current tenant
3. Auto-injects tenant ID on record creation
4. Shows tenant selector in the header

```yaml
multitenant:
  enabled: true
  resolver: subdomain          # subdomain | header
  require_isolated_db: true    # Prevents tenants sharing DBs
  database_alias_template: "tenant_%s"
  tenants:
    acme:
      site: main
      database: tenant_acme
    globex:
      site: main
      database: tenant_globex
```

**Model Declaration:**

```go
type Order struct {
    model.BaseModel
    ID        uint      `db:"id;pk"`
    TenantID  string    `db:"tenant;column:tenant_id"` // Tenant field
    Product   string    `db:"column:product"`
    Quantity  int       `db:"column:quantity"`
}
```

### Multi-Site Configuration

When `multisite.enabled: true`, the admin panel shows site information:

```yaml
multisite:
  enabled: true
  default_site: main
  sites:
    main:
      hosts: ["example.com", "www.example.com"]
      database: default
    eu:
      hosts: ["eu.example.com"]
      database: eu_db
```

### RBAC Configuration

To enable fine-grained access control:

```yaml
admin_rbac_policy_file: admin_rbac.csv
```

**RBAC Policy File Format (CSV):**

```csv
p, admin, admin:*, *
p, editor, admin:User, read
p, editor, admin:User, update
p, viewer, admin:User, read
g, alice, admin
g, bob, editor
g, carol, viewer
```

**Policy Syntax:**
- `p, subject, object, action` - Permission policy
- `g, user, role` - Role assignment

**Objects:** `admin:<ModelName>` or `admin:*` for all models
**Actions:** `read`, `create`, `update`, `delete`, `list_models`, `system_pulse`, `live_traffic`, `list_sessions`, `rbac_list`, `rbac_manage`, `audit_view`, `audit_manage`, `migration_view`, `migration_apply`, `health_check`, `jobs_view`, `sites_view`, `deployment_view`, `cache_view`, `cache_manage`, `storage_view`, `email_view`

### Audit Logging

Audit logging is enabled by default with an in-memory bounded store (10,000 entries).

To disable or configure:

```yaml
# Currently hardcoded; future versions may expose config
# AuditEnabled: false  # Disable audit logging
# AuditMaxSize: 50000  # Increase max entries
```

## Navigation Structure

| Section | Hash Route | Description |
|---------|-----------|-------------|
| Overview | `#/` | Model listing, database groups, record counts |
| Data Studio | `#/data-studio` | CRUD browser with search/filter/sort |
| System Pulse | `#/system` | Go runtime, DB pools, feature flags, queues |
| Network Inspector | `#/live` | Live HTTP traffic, SQL queries, WebSocket stream |
| Infra Manager | `#/sessions` | Active sessions with time-series charts |
| Health | `#/health` | Database connectivity checks (Redis: URL-only, no ping) |
| Access Control | `#/rbac` | RBAC policies and role management |
| Audit Log | `#/audit` | CRUD operation audit trail |
| Migrations | `#/migrations` | Migration file listing (apply via `goframe migrate` CLI) |
| Deployment | `#/deployment` | Runtime info, cluster topology |
| Jobs | `#/jobs` | Background job queue status |
| Cache | `#/cache` | Redis URL presence check (detailed stats: use `redis-cli`) |
| Storage | `#/storage` | Local filesystem browser (cloud providers: future) |
| Sites | `#/sites` | Multi-site listing |

## API Endpoints

All endpoints are under `/admin/api/`.

### Models
| Method | Path | Description |
|--------|------|-------------|
| GET | `/models?stats=full|light` | List all models with counts |
| GET | `/models/{name}/schema` | Get model field schema |
| GET | `/models/{name}?page=&page_size=&search=&db=` | List records |
| POST | `/models/{name}?db=` | Create record |
| GET | `/models/{name}/{id}?db=` | Get record |
| PUT | `/models/{name}/{id}?db=` | Update record |
| DELETE | `/models/{name}/{id}?db=` | Delete record |
| POST | `/models/{name}/bulk` | Bulk actions (delete/export) |
| GET | `/models/{name}/export?db=` | CSV export |

### RBAC
| Method | Path | Description |
|--------|------|-------------|
| GET | `/rbac/policies` | List all policies and roles |
| POST | `/rbac/policies` | Add policy (sub, obj, act) |
| DELETE | `/rbac/policies` | Remove policy |
| POST | `/rbac/roles/assign` | Assign role to user |
| POST | `/rbac/roles/remove` | Remove role from user |
| GET | `/rbac/roles?user=` | Get user roles |
| GET | `/rbac/check?sub=&obj=&act=` | Check permission |

### Audit
| Method | Path | Description |
|--------|------|-------------|
| GET | `/audit?page=&page_size=&user_id=&model=&action=` | List audit entries |
| POST | `/audit/clear` | Clear audit log |

### Health
| Method | Path | Description |
|--------|------|-------------|
| GET | `/health` | System health check |

### Management
| Method | Path | Description |
|--------|------|-------------|
| GET | `/migrations` | List migration files and status |
| POST | `/migrations/apply?steps=` | Count pending; apply via `goframe migrate` CLI |
| GET | `/jobs` | Background job status |
| GET | `/sites` | List sites |
| GET | `/deployment` | Deployment info |
| GET | `/cache` | Redis URL presence (stub: use `redis-cli` for details) |
| POST | `/cache/flush` | Stub: use `redis-cli FLUSHDB` |
| GET | `/storage?path=` | Local filesystem listing (cloud: future) |
| GET | `/email` | Stub: requires mail driver integration |

## Deployment Scenarios

### Standalone

Single process, in-memory sessions, SQLite default. All features work locally.

### Docker Cluster

Multiple containers behind load balancer with Redis for shared sessions:

```yaml
session_store: redis
session_redis_url: redis://redis:6379/0
admin_cluster_enabled: true
admin_cluster_redis_url: redis://redis:6379/1
```

The admin panel aggregates telemetry across nodes via Redis pub/sub.

### Kubernetes

Deploy as a Deployment/StatefulSet. The admin detects pod metadata automatically:

- Pod name, namespace, node
- Cluster topology via Redis relay
- Health checks integrate with K8s readiness/liveness probes

**Health probe endpoint:** `/admin/api/health`

### Behind Load Balancer

Use shared Redis session store to maintain sessions across requests that may hit different nodes:

```yaml
session_store: redis
admin_cluster_enabled: true
```

## Security Considerations

1. **Authentication**: Session-based via `goframe_admin_users` table
2. **Authorization**: RBAC via Casbin (optional but recommended for multi-user environments)
3. **Tenant Isolation**: Auto-filtered when multi-tenant enabled
4. **Audit Trail**: All CRUD operations logged automatically
5. **Environment Variables**: Masked in System Pulse (KEY, SECRET, PASSWORD, TOKEN)
6. **SQL Arguments**: Redacted in Network Inspector
7. **Payload Previews**: Censored for sensitive data

### Storage Configuration

The admin panel's file storage browser uses the GoFrame storage layer (`pkg/storage`).
Configure your storage provider:

```yaml
storage:
  provider: s3                # s3 | gcs | azure | local
  s3:
    endpoint: ""              # Empty = AWS S3
    bucket: myapp-files
    region: us-east-1
    access_key_id: "${AWS_ACCESS_KEY_ID}"
    secret_access_key: "${AWS_SECRET_ACCESS_KEY}"
  local:
    path: storage/            # Development only
```

For public file serving, configure path mapping:

```yaml
storage:
  public_url_base: "https://cdn.example.com"
  public_paths:
    /media: storage/public/media/
    /assets: storage/public/assets/
```

The framework automatically mounts HTTP routes for public paths, so files stored
at `storage/public/media/blog/hero.png` are served at `https://cdn.example.com/media/blog/hero.png`.

## Troubleshooting

### Admin panel not accessible
- Check `admin_prefix` configuration
- Verify database has `goframe_admin_users` table
- Check bootstrap credentials in logs

### RBAC not working
- Verify `admin_rbac_policy_file` path exists
- Check CSV format (p = policy, g = grouping/role)
- Superusers bypass RBAC policies

### Tenant filtering not working
- Ensure model has tenant field (`db:"tenant"` tag or `tenant_id` column)
- Verify `multitenant.enabled: true`
- Check tenant resolution (subdomain vs header)

### Cluster telemetry not aggregating
- Verify Redis URL configured for cluster relay
- Check `admin_cluster_enabled: true`
- Ensure nodes share the same cluster channel
