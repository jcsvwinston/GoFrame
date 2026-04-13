# Admin UI Documentation

Reference date: 2026-04-13.
Status: Current (v0.7.x - React-based UI).

## Overview

The GoFrame admin panel (`/admin`) is an embedded **React + TypeScript** application with Tailwind CSS and shadcn/ui components. It provides:

- Modern authentication UI with dark/light theme support
- Data management (CRUD) for registered models
- Real-time runtime inspection (traffic, SQL, sessions)
- System health monitoring
- Multi-tenant and multi-site management
- Role-based access control (RBAC)
- Audit logging
- Background job monitoring

## Architecture

### Technology Stack

| Component | Technology |
|-----------|-----------|
| **Framework** | React 19 (TypeScript) |
| **Bundler** | Vite 6 |
| **Styling** | Tailwind CSS 3 |
| **UI Components** | shadcn/ui (local) |
| **State Management** | Zustand 5 |
| **Routing** | React Router 7 |
| **Charts** | Recharts 2 |
| **Icons** | Lucide React |

### Project Structure

```
pkg/admin/ui/
├── src/
│   ├── components/
│   │   ├── ui/             # shadcn/ui components
│   │   └── layout/         # Dashboard layout
│   ├── features/           # Feature modules
│   │   ├── auth/           # Login page
│   │   ├── overview/       # Dashboard
│   │   ├── data-studio/    # Export/Import
│   │   ├── system/         # System metrics
│   │   ├── network/        # Live traffic
│   │   ├── infra/          # Sessions
│   │   ├── health/         # Health checks
│   │   ├── rbac/           # Access control
│   │   └── audit/          # Audit log
│   ├── services/           # API integration
│   ├── stores/             # Zustand stores
│   ├── types/              # TypeScript types
│   └── lib/                # Utilities
├── package.json
├── vite.config.ts
├── tailwind.config.js
└── README.md
```

## Development

### Prerequisites

- Node.js 18+
- npm 9+

### Setup

```bash
cd pkg/admin/ui
npm install
```

### Development Server

```bash
npm run dev
```

Starts Vite dev server at `http://localhost:5173` with hot reload.

### Production Build

```bash
# From admin directory
./build-ui.sh

# Or manually
cd pkg/admin/ui
npm run build
```

Output: `ui/dist/` (embedded in Go binary via `//go:embed`)

## Login Page

### Features

- Modern card-based design with gradient background
- Theme toggle (dark/light mode)
- Form validation with real-time feedback
- Toast notifications for success/error
- Loading states during authentication
- Responsive design (mobile-friendly)

### Authentication Flow

1. **GET** `/admin/login` → React login page rendered
2. **POST** `/admin/login` → Go validates credentials against `goframe_admin_users` table
3. **Success** → Session created, redirect to `/admin/`
4. **Failure** → React login page shown with error toast

### Session Management

Sessions are stored in the configured backend (`memory|sql|redis`) with keys:
- `__goframe_admin_user_id`
- `__goframe_admin_username`
- `__goframe_admin_email`
- `__goframe_admin_superuser`

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

## Dashboard Modules

### 1. Overview (`/`)
- Model statistics with record counts
- Quick actions panel
- System information
- Database connection status

### 2. Data Studio (`/data-studio`)
- Export data (CSV, JSON, SQL formats)
- Import data with file upload
- Format selector
- Progress indicators

### 3. System Pulse (`/system`)
- Go runtime metrics (goroutines, memory, GC)
- Real-time updates (5-second interval)
- Charts with Recharts
- Database pool status

### 4. Network Inspector (`/live`)
- Live HTTP traffic monitoring
- WebSocket integration for real-time updates
- Request log with method/status badges
- Start/Stop/Clear controls

### 5. Session Manager (`/sessions`)
- Active sessions table
- User agent and IP display
- Session termination capability
- Refresh functionality

### 6. Health Checks (`/health`)
- Service health status
- Healthy/Unhealthy counts
- Latency display
- Status badges with colors

### 7. Access Control (`/rbac`)
- Policy list table
- Add policy dialog
- Delete policy functionality
- Role/Resource/Action management

### 8. Audit Log (`/audit`)
- Audit trail table
- Search functionality
- Action badges with colors
- Timestamp and user display

## Customization

### Modify Theme Colors

Edit `pkg/admin/ui/tailwind.config.js`:

```javascript
theme: {
  extend: {
    colors: {
      primary: {
        DEFAULT: '#your-color',
        foreground: '#your-text-color',
      }
    }
  }
}
```

### Add New UI Components

```bash
cd pkg/admin/ui
npx shadcn@latest add tooltip
# Installs to src/components/ui/tooltip.tsx
```

### Create New Feature Pages

```bash
mkdir -p src/features/my-feature/pages
# Follow existing patterns in other features/
```

### Extend API Integration

Edit `src/services/api.ts` and add new endpoint functions.

## Security Considerations

1. **Authentication**: Session-based via `goframe_admin_users` table
2. **Authorization**: RBAC via Casbin (optional but recommended for multi-user environments)
3. **Tenant Isolation**: Auto-filtered when multi-tenant enabled
4. **Audit Trail**: All CRUD operations logged automatically
5. **Environment Variables**: Masked in System Pulse (KEY, SECRET, PASSWORD, TOKEN)
6. **SQL Arguments**: Redacted in Network Inspector
7. **Payload Previews**: Censored for sensitive data
8. **Zero CDN Dependencies**: All packages installed locally

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

### Build fails
- Run `npm install` in `pkg/admin/ui/`
- Check Node.js version (18+ required)
- Clear `node_modules` and reinstall: `rm -rf node_modules && npm install`

### UI not updating after rebuild
- Ensure `npm run build` completed successfully
- Restart Go application to reload embedded assets
- Check browser cache (hard refresh: Ctrl+Shift+R / Cmd+Shift+R)

## Build Integration

### Go Embed Configuration

**Login Page** (`default_auth.go`):
```go
//go:embed all:ui/dist/*
var loginUIFS embed.FS
```

**Dashboard** (`panel.go`):
```go
//go:embed all:ui/dist/*
var uiFS embed.FS
```

### Build Process

Both login page and dashboard are served from the same React build:

```bash
npm run build  # Creates ui/dist/
# Output embedded in Go binary at compile time
```

**Important:**
- Build must run before Go compilation
- Development files (`node_modules`, `src`) NOT embedded
- Only `ui/dist/` included in binary

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

### Export/Import
| Method | Path | Description |
|--------|------|-------------|
| POST | `/export` | Create export |
| GET | `/export/list` | List exports |
| GET | `/export/status` | Export status |
| GET | `/export/download` | Download export |
| POST | `/import/upload` | Upload import file |
| POST | `/import/validate` | Validate import |
| POST | `/import/execute` | Execute import |

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

### System
| Method | Path | Description |
|--------|------|-------------|
| GET | `/system/snapshot` | Go runtime metrics |
| GET | `/system/flags` | List feature flags |
| POST | `/system/flags` | Create feature flag |
| PUT | `/system/flags/{name}` | Update feature flag |
| DELETE | `/system/flags/{name}` | Delete feature flag |

### Live Traffic
| Method | Path | Description |
|--------|------|-------------|
| GET | `/live/snapshot` | Current live requests |
| GET | `/live/ws` | WebSocket stream |
| GET | `/live/excludes` | List exclude patterns |
| POST | `/live/excludes` | Add exclude pattern |
| DELETE | `/live/excludes` | Remove exclude pattern |

### Management
| Method | Path | Description |
|--------|------|-------------|
| GET | `/sessions` | List active sessions |
| GET | `/health` | System health check |
| GET | `/migrations` | List migration files and status |
| POST | `/migrations/apply?steps=` | Apply migrations |
| GET | `/jobs` | Background job status |
| GET | `/sites` | List sites |
| GET | `/deployment` | Deployment info |
| GET | `/cache` | Redis URL presence |
| POST | `/cache/flush` | Flush cache |
| GET | `/storage?path=` | Local filesystem listing |
| GET | `/email` | Mail stats |

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

## Architecture & Design

### Core Constraints

1. **No massive telemetry persistence** - Historical/long-retention telemetry belongs to external OpenTelemetry backends. Admin uses in-memory data only.

2. **Zero-overhead by default** - Collectors are non-blocking for request/DB hot paths. When no admin consumer exists, events are dropped quickly.

3. **Security by default** - `/admin/*` is auth-protected. Sensitive environment values are masked. Payload logging supports redaction/censoring.

4. **Durable contracts** - Add new capabilities without breaking stable APIs/CLI/config contracts. Admin surface moves from `transitional` to `stable` only after test-backed acceptance criteria are met.

### Implementation Patterns

- **Observer + middleware pattern**: Standard `net/http` middleware hooks for instrumentation
- **Non-blocking event pipeline**: Bounded channels + drop policy + lightweight fanout
- **Secure routing envelope**: Mounted under `/admin/*` with auth middleware
- **UI isolation**: Self-contained admin assets to avoid collisions with app assets

### Live Traffic Architecture

**Active Sessions Tracker:**
- Sessions stored in `sync.Map`
- Shows user id, IP, user-agent, last route, last seen

**Identity to Trace Mapping:**
- Shows `trace_id` associated with current session/request context

**Live SQL Sniffer:**
- Non-blocking event capture around DB execution
- SQL text, args (redacted), duration (ms), status/error
- Dispatch to connected admin WebSocket subscribers only

**Live Request/Response Watcher:**
- Bounded ring buffer with last N HTTP events
- Route, status code, duration, censored payload preview

### System Pulse Architecture

**Goroutine Explorer:**
- Surface `pprof` goroutine snapshot and grouped state counts

**DB Pool Stats:**
- Expose `db.Stats()` live values (`InUse`, `Idle`, `WaitCount`, etc.)

**Memory and GC:**
- Expose selected `runtime.ReadMemStats` counters and trends

**Environment/Config Viewer:**
- Static snapshot from startup/runtime config
- Auto-mask keys containing `KEY`, `SECRET`, `PASSWORD`, `TOKEN`

**Worker/Job Pool Monitor:**
- Expose active workers, queued jobs, in-progress jobs (when task runtime is enabled)

**Live Feature Flags:**
- Runtime boolean flags toggled in-memory without restart
- Include audit metadata in memory/log stream
