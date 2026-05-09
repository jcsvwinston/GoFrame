# Authentication & Authorization Guide

Reference date: 2026-04-10.
Status: Current.

This guide covers GoFrame's authentication (`pkg/auth`) and authorization (`pkg/authz`) systems, including JWT flows, session management, password handling, and Casbin-backed policy enforcement.

## Table of Contents

- [Overview](#overview)
- [Authentication (`pkg/auth`)](#authentication-pkgauth)
  - [JWT Authentication](#jwt-authentication)
  - [Server-Side Sessions](#server-side-sessions)
  - [Password Hashing](#password-hashing)
  - [User Model](#user-model)
- [Authorization (`pkg/authz`)](#authorization-pkgauthz)
  - [Casbin Integration](#casbin-integration)
  - [Policy Files](#policy-files)
  - [Authorization Middleware](#authorization-middleware)
- [Admin Authentication Flow](#admin-authentication-flow)
- [Production Checklist](#production-checklist)

---

## Overview

GoFrame provides two complementary authentication mechanisms:

| Mechanism | Best For | Storage Options |
|-----------|----------|-----------------|
| **JWT** | Stateless APIs, mobile apps, microservices | None (token-only) |
| **Sessions** | Server-rendered apps, admin panel, web UIs | Memory, SQL, Redis |

Authorization is handled separately by `pkg/authz`, which integrates with Casbin for policy-based access control.

---

## Authentication (`pkg/auth`)

### JWT Authentication

JWT (JSON Web Token) authentication is ideal for stateless API endpoints.

#### Configuration

```yaml
# goframe.yaml
jwt_secret: your-super-secret-key-change-in-production
jwt_expiry: 24h
jwt_issuer: myapp
```

#### Generating Tokens

```go
import "github.com/jcsvwinston/nucleus/pkg/auth"

// Create a JWT manager
manager := auth.NewJWTManager(cfg.JWTSecret, cfg.JWTExpiry, cfg.JWTIssuer)

// Generate token for a user
token, err := manager.GenerateToken(auth.JWTClaims{
    UserID:   "user-123",
    Email:    "alice@example.com",
    Role:     "admin",
    CustomClaims: map[string]any{
        "tenant_id": "acme",
    },
})
if err != nil {
    return err
}

// Return to client
return ctx.JSON(200, map[string]string{"access_token": token})
```

#### Validating Tokens (Middleware)

GoFrame's router includes JWT middleware that validates tokens and enriches the request context:

```go
import "github.com/jcsvwinston/nucleus/pkg/router"

r := router.New()

// Protect routes with JWT middleware
r.Use(router.JWTMiddleware(manager))

// Handlers can access user context
r.GET("/api/profile", func(w http.ResponseWriter, r *http.Request) {
    userID := observe.UserIDFromCtx(r.Context())
    requestID := observe.RequestIDFromCtx(r.Context())
    traceID := observe.TraceIDFromCtx(r.Context())

    // JWT middleware enriches context with these values
    ctx := r.Context()
    // ... use context values
})
```

#### Token Refresh Flow

For security-sensitive applications, implement short-lived access tokens with refresh tokens:

```go
// Access token: short-lived (15m)
accessCfg := auth.JWTConfig{
    Secret: cfg.JWTSecret,
    Expiry: 15 * time.Minute,
    Issuer: "myapp",
}

// Refresh token: long-lived (7d), stored server-side
refreshCfg := auth.JWTConfig{
    Secret: cfg.JWTSecret,
    Expiry: 7 * 24 * time.Hour,
    Issuer: "myapp",
}

// Generate both tokens on login
accessToken, _ := accessManager.GenerateToken(claims)
refreshToken, _ := refreshManager.GenerateToken(auth.JWTClaims{
    UserID: claims.UserID,
    CustomClaims: map[string]any{
        "token_type": "refresh",
        "jti": generateUniqueID(), // Store in DB for revocation
    },
})
```

#### Token Revocation

JWTs are stateless by design. To revoke tokens before expiry:

1. **Short expiry + refresh tokens**: Use 15-minute access tokens; revoke refresh tokens server-side.
2. **Token blacklist**: Store revoked JWT IDs (`jti`) in Redis with TTL matching token expiry.
3. **Version field**: Add a `token_version` claim; increment on password change/logout.

### Server-Side Sessions

Sessions are required for server-rendered applications, admin panel, and CSRF-protected forms.

#### Configuration

```yaml
# goframe.yaml
session_store: sql          # Options: memory, sql, redis
session_cookie_name: goframe_session
session_cookie_secure: true # Set true in production (HTTPS only)
session_cookie_http_only: true
session_cookie_same_site: strict
session_idle_timeout: 30m
session_table: goframe_sessions
```

#### Store Backends

| Store | Use Case | Configuration |
|-------|----------|---------------|
| **Memory** | Development, single-instance testing | `session_store: memory` |
| **SQL** | Production, multi-replica without Redis | `session_store: sql`, `session_table: goframe_sessions` |
| **Redis** | High-scale, distributed sessions | `session_store: redis`, `session_redis_url: redis://localhost:6379/0` |

#### Session Usage

```go
import "github.com/jcsvwinston/nucleus/pkg/auth"

// In handler (session middleware wired by app.New)
session := auth.SessionFromContext(r.Context())

// Set session data
session.Put(r.Context(), "user_id", "user-123")
session.Put(r.Context(), "role", "admin")

// Get session data
userID := session.GetString(r.Context(), "user_id")
if userID == "" {
    http.Redirect(w, r, "/login", http.StatusSeeOther)
    return
}

// Destroy session (logout)
session.Destroy(r.Context())
```

#### Session Runtime Metadata

GoFrame automatically enriches sessions with serving-node identity for cluster diagnostics:

```go
// Session metadata includes:
// - first_seen: When session was created
// - last_seen: Last activity timestamp
// - pod: Pod/container identifier
// - host: Hostname
// - instance: Process instance ID
```

View active sessions via admin UI at `/admin#/sessions` or API at `GET /admin/api/sessions`.

#### Session Maintenance Commands

```bash
# Create SQL session table (if using sql store)
goframe createcachetable --config goframe.yaml

# Clear expired sessions (production-safe)
goframe clearsessions --config goframe.yaml

# Clear all sessions (use with caution)
goframe clearsessions --all --force --config goframe.yaml
```

### Password Hashing

GoFrame uses `bcrypt` for password hashing via `golang.org/x/crypto/bcrypt`.

```go
import "github.com/jcsvwinston/nucleus/pkg/auth"

// Hash a password (use default cost=10 in production)
hash, err := auth.HashPassword("plaintext_password")
if err != nil {
    return err
}

// Verify a password
isValid := auth.CheckPasswordHash("plaintext_password", hash)
if !isValid {
    return fmt.Errorf("invalid credentials")
}

// Custom cost (higher = slower, more secure)
hash, err = auth.HashPasswordWithCost("password", bcrypt.MaxCost)
```

**Security recommendations:**

- Never log or store plaintext passwords.
- Use bcrypt default cost (10) for most applications.
- Increase cost for high-security applications (cost 12-14).
- Implement rate limiting on login endpoints.

### User Model

GoFrame provides a minimal user structure in `pkg/auth`:

```go
type User struct {
    ID       string
    Username string
    Email    string
    Role     string
    HashedPassword string
    CreatedAt time.Time
    UpdatedAt time.Time
}
```

Admin users are managed via CLI commands:

```bash
# Create admin user
goframe createuser --config goframe.yaml --username admin --email admin@example.com

# Interactive password prompt
goframe createuser --config goframe.yaml --username admin

# Non-interactive (CI/CD)
goframe createuser --config goframe.yaml --username admin --password "secure-password" --no-input

# Change password
goframe changepassword --config goframe.yaml --username admin
goframe changepassword --config goframe.yaml --username admin --password "new-password" --no-input
```

---

## Authorization (`pkg/authz`)

### Casbin Integration

GoFrame integrates with [Casbin](https://casbin.org/) for policy-based authorization.

#### Configuration

```yaml
# goframe.yaml
authz_model_path: internal/config/authz_model.conf
authz_policy_path: internal/config/authz_policy.csv
```

#### Model File (`authz_model.conf`)

Define your access control model:

```ini
[request_definition]
r = sub, obj, act

[policy_definition]
p = sub, obj, act

[policy_effect]
e = some(where (p.eft == allow))

[matchers]
m = r.sub == p.sub && keyMatch(r.obj, p.obj) && regexMatch(r.act, p.act)
```

#### Policy File (`authz_policy.csv`)

Define your policies:

```csv
p, admin, /admin/*, *
p, admin, /api/*, *
p, editor, /api/articles, POST
p, editor, /api/articles/*, PUT
p, editor, /api/articles/*, DELETE
p, viewer, /api/*, GET
p, anonymous, /api/health, GET
p, anonymous, /login, GET
p, anonymous, /login, POST
```

#### Enforcer Usage

```go
import "github.com/jcsvwinston/nucleus/pkg/authz"

// Initialize enforcer
enforcer, err := authz.NewEnforcer(cfg.AuthzModelPath, cfg.AuthzPolicyPath)
if err != nil {
    return err
}

// Check permissions
allowed, err := enforcer.Enforce("alice", "/api/articles", "GET")
if err != nil {
    return err
}
if !allowed {
    return fmt.Errorf("forbidden")
}
```

#### Authorization Middleware

Apply authorization middleware to routes:

```go
import "github.com/jcsvwinston/nucleus/pkg/authz"

// Role-based middleware
r.Use(authz.RoleMiddleware(enforcer, "admin", "editor"))

// Custom middleware with dynamic subject
r.Use(func(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        role := observe.UserIDFromCtx(r.Context()) // Or extract from JWT claims
        allowed, _ := enforcer.Enforce(role, r.URL.Path, r.Method)
        if !allowed {
            http.Error(w, "Forbidden", http.StatusForbidden)
            return
        }
        next.ServeHTTP(w, r)
    })
})
```

#### Policy Management

```go
// Add policy at runtime
enforcer.AddPolicy("editor", "/api/drafts", "POST")

// Remove policy
enforcer.RemovePolicy("editor", "/api/drafts", "POST")

// Check if policy exists
hasPolicy := enforcer.HasPolicy("admin", "/admin/*", "*")

// Get all policies
policies := enforcer.GetPolicy()
```

---

## Admin Authentication Flow

The admin panel has two authentication modes:

### Bootstrap Mode

When there are **no rows** in `goframe_admin_users`, `/admin` is accessible without login to help with initial setup.

```bash
# First access after install - no authentication required
# Open http://localhost:8080/admin
```

### Protected Mode

Once at least one admin user exists, `/admin` requires login at `/admin/login`.

```bash
# Create first admin user
goframe createuser --config goframe.yaml --username admin --email admin@example.com

# All subsequent accesses require login
# Login at http://localhost:8080/admin/login
```

### Admin Session Security

- Admin sessions use the configured session store (`sql` or `redis` recommended for production).
- Session cookies are marked `HttpOnly`, `Secure` (in production), and `SameSite=Strict`.
- View active admin sessions at `/admin/api/sessions` or UI at `/admin#/sessions`.

---

## Production Checklist

- [ ] Set strong `jwt_secret` (use random 64-byte hex key).
- [ ] Use `session_store: redis` or `sql` for multi-replica deployments.
- [ ] Set `session_cookie_secure: true` when using HTTPS.
- [ ] Set `session_cookie_same_site: strict` for CSRF protection.
- [ ] Implement rate limiting on login endpoints (`rate_limit_by_route` or `rate_limit_burst`).
- [ ] Use Casbin policies for fine-grained authorization.
- [ ] Store `authz_policy.csv` in version control; reload on changes.
- [ ] Run `goframe clearsessions` on a cron schedule to clean expired sessions.
- [ ] Monitor admin session dashboard for unusual access patterns.
- [ ] Rotate `jwt_secret` periodically (requires token re-issuance).
