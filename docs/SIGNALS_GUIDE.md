# Signals & Event Bus Guide

Reference date: 2026-04-10.
Status: Current.

This guide covers GoFrame's in-process event bus (`pkg/signals`), used for model hooks, domain events, and cross-cutting communication within the application.

## Table of Contents

- [Overview](#overview)
- [Publishing Events](#publishing-events)
- [Subscribing to Events](#subscribing-to-events)
- [Model Hook Integration](#model-hook-integration)
- [Event Ordering](#event-ordering)
- [Error Handling in Subscribers](#error-handling-in-subscribers)
- [Use Cases](#use-cases)

---

## Overview

GoFrame's signals package provides a lightweight, in-process pub/sub system for:

- **Model lifecycle hooks**: `before_save`, `after_create`, `after_delete`
- **Domain events**: `user_registered`, `order_completed`, `payment_received`
- **System events**: `app_started`, `shutdown_requested`

The event bus is **synchronous by default** — subscribers are called in registration order within the same goroutine.

---

## Publishing Events

```go
import "github.com/jcsvwinston/GoFrame/pkg/signals"

// Define event signals
var (
    UserRegistered = signals.NewSignal()
    OrderCompleted = signals.NewSignal()
    ArticlePublished = signals.NewSignal()
)

// Publish an event with payload
err := UserRegistered.Send(map[string]any{
    "user_id": "user-123",
    "email":   "alice@example.com",
    "source":  "web",
})
```

### Send with context

```go
ctx := context.Background()
ctx = observe.CtxWithRequestID(ctx, "req-abc")

err := UserRegistered.SendCtx(ctx, map[string]any{
    "user_id": "user-123",
})
```

---

## Subscribing to Events

### Basic subscription

```go
// Subscribe with a handler
UserRegistered.Connect(func(payload any) error {
    data, ok := payload.(map[string]any)
    if !ok {
        return fmt.Errorf("unexpected payload type")
    }

    userID := data["user_id"].(string)
    log.Printf("User registered: %s", userID)

    // Send welcome email
    return sendWelcomeEmail(userID)
})
```

### Multiple subscribers

```go
// Subscriber 1: Send welcome email
UserRegistered.Connect(func(payload any) error {
    return sendWelcomeEmail(payload)
})

// Subscriber 2: Create analytics record
UserRegistered.Connect(func(payload any) error {
    return analytics.Track("user_registered", payload)
})

// Subscriber 3: Notify admin channel
UserRegistered.Connect(func(payload any) error {
    return notifySlack("#signups", payload)
})
```

Subscribers are called **in registration order**.

---

## Model Hook Integration

GoFrame models can integrate with signals for lifecycle events:

```go
type Article struct {
    model.BaseModel
    Title  string
    Slug   string
    Status string
}

func (a *Article) BeforeSave() error {
    // Generate slug if not set
    if a.Slug == "" {
        a.Slug = slug.Make(a.Title)
    }

    // Emit signal for external listeners
    if a.ID == 0 { // New record
        return ArticlePublished.Send(map[string]any{
            "article_id": a.ID,
            "title":      a.Title,
            "author_id":  a.AuthorID,
        })
    }

    return nil
}

func (a *Article) AfterDelete() error {
    return signals.NewSignal("article_deleted").Send(map[string]any{
        "article_id": a.ID,
    })
}
```

### Common model signals

| Signal | When Emitted | Typical Use |
|--------|-------------|-------------|
| `before_save` | Before INSERT or UPDATE | Slug generation, timestamp updates |
| `after_create` | After INSERT | Send notifications, update counters |
| `after_update` | After UPDATE | Invalidate cache, audit log |
| `after_delete` | After DELETE | Clean up related resources |

---

## Event Ordering

### Guarantee: Registration order

Subscribers are called in the order they were connected:

```go
// This subscriber runs first
UserRegistered.Connect(func(payload any) error {
    fmt.Println("1. Creating user record")
    return nil
})

// This subscriber runs second
UserRegistered.Connect(func(payload any) error {
    fmt.Println("2. Sending welcome email")
    return nil
})

// This subscriber runs third
UserRegistered.Connect(func(payload any) error {
    fmt.Println("3. Notifying analytics")
    return nil
})
```

### No parallelism guarantee

Subscribers run **sequentially** in the same goroutine that calls `Send()`. If a subscriber blocks, all subsequent subscribers wait.

For async processing, delegate to a worker:

```go
UserRegistered.Connect(func(payload any) error {
    // Enqueue background job instead of blocking
    return taskManager.EnqueueJSON("emails.send_welcome", payload)
})
```

---

## Error Handling in Subscribers

### Error propagation

If a subscriber returns an error, `Send()` returns immediately with that error:

```go
UserRegistered.Connect(func(payload any) error {
    err := sendWelcomeEmail(payload)
    if err != nil {
        return fmt.Errorf("welcome email failed: %w", err)
    }
    return nil
})

// Send returns the error from the failing subscriber
err := UserRegistered.Send(payload)
if err != nil {
    log.Error("user registration handler failed", "error", err)
    // Subsequent subscribers were NOT called
}
```

### Continue-on-error pattern

If you want all subscribers to run regardless of individual failures:

```go
func ConnectResilient(signal *signals.Signal, handler func(any) error) {
    signal.Connect(func(payload any) error {
        err := handler(payload)
        if err != nil {
            log.Warn("subscriber error (non-blocking)", "error", err)
            return nil // Swallow error, continue to next subscriber
        }
        return nil
    })
}

// Usage
ConnectResilient(UserRegistered, func(payload any) error {
    return analytics.Track("user_registered", payload)
})
```

### Timeout for subscribers

```go
UserRegistered.Connect(func(payload any) error {
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    return handlerWithTimeout(ctx, payload)
})
```

---

## Use Cases

### Audit logging

```go
var AuditLog = signals.NewSignal()

func init() {
    AuditLog.Connect(func(payload any) error {
        entry, ok := payload.(AuditEntry)
        if !ok {
            return fmt.Errorf("invalid audit entry")
        }
        return db.Exec(
            "INSERT INTO audit_log (action, user_id, details, created_at) VALUES (?, ?, ?, ?)",
            entry.Action, entry.UserID, entry.Details, time.Now(),
        )
    })
}

// Emit from handlers
AuditLog.Send(AuditEntry{
    Action: "user_login",
    UserID: userID,
    Details: map[string]any{"ip": r.RemoteAddr},
})
```

### Cache invalidation

```go
ArticlePublished.Connect(func(payload any) error {
    data := payload.(map[string]any)
    articleID := data["article_id"].(string)

    // Invalidate article cache
    cache.Delete("article:" + articleID)
    cache.Delete("articles:list")
    return nil
})
```

### Webhook dispatch

```go
OrderCompleted.Connect(func(payload any) error {
    return taskManager.EnqueueJSON("webhooks.deliver", map[string]any{
        "event":  "order.completed",
        "payload": payload,
    })
})
```

### System startup/shutdown

```go
var AppStarted = signals.NewSignal()
var AppStopping = signals.NewSignal()

func init() {
    AppStarted.Connect(func(payload any) error {
        log.Info("application started", "version", version)
        return nil
    })

    AppStopping.Connect(func(payload any) error {
        log.Info("application shutting down")
        return nil
    })
}

// In main.go
func main() {
    app, _ := app.New(cfg)
    AppStarted.Send(nil)
    defer AppStopping.Send(nil)
    app.Run(context.Background())
}
```

---

## API Reference

```go
// Create a new signal
sig := signals.NewSignal()

// Subscribe
sig.Connect(handler func(payload any) error)

// Publish (sync, blocks until all subscribers complete)
err := sig.Send(payload)

// Publish with context
err := sig.SendCtx(ctx, payload)

// Get subscriber count
count := sig.ListenerCount()

// Disconnect all subscribers (for testing)
sig.Reset()
```
