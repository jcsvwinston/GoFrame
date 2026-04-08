package admin

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/jcsvwinston/GoFrame/pkg/db"
	"github.com/jcsvwinston/GoFrame/pkg/model"
	"github.com/jcsvwinston/GoFrame/pkg/observe"
)

func TestRequestRingBufferLatest(t *testing.T) {
	ring := newRequestRingBuffer(3)
	ring.push(liveRequestEvent{Path: "/a"})
	ring.push(liveRequestEvent{Path: "/b"})
	ring.push(liveRequestEvent{Path: "/c"})
	ring.push(liveRequestEvent{Path: "/d"})

	rows := ring.latest(10)
	if len(rows) != 3 {
		t.Fatalf("expected 3 rows, got %d", len(rows))
	}
	if rows[0].Path != "/d" || rows[1].Path != "/c" || rows[2].Path != "/b" {
		t.Fatalf("unexpected order: %#v", rows)
	}
}

func TestLiveEventBusDropsWhenSubscriberIsSlow(t *testing.T) {
	bus := newLiveEventBus(1)
	_, unsubscribe := bus.subscribe()
	defer unsubscribe()

	bus.publish(liveEventEnvelope{Type: "a"})
	bus.publish(liveEventEnvelope{Type: "b"})
	bus.publish(liveEventEnvelope{Type: "c"})

	stats := bus.stats()
	if stats.Published != 3 {
		t.Fatalf("expected published=3, got %d", stats.Published)
	}
	if stats.Dropped == 0 {
		t.Fatalf("expected dropped > 0, got %d", stats.Dropped)
	}
}

func TestLiveTrafficMiddlewareRecordsRequestAndSession(t *testing.T) {
	panel, cleanup := setupPanelForTest(t, db.EngineSQL)
	defer cleanup()

	mw := panel.liveTrafficMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
	}))

	req := httptest.NewRequest(http.MethodGet, "/products?token=abc123&name=john", nil)
	ctx := observe.CtxWithRequestID(req.Context(), "req-1")
	ctx = observe.CtxWithTraceID(ctx, "trace-1")
	ctx = observe.CtxWithUserID(ctx, "user-42")
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	mw.ServeHTTP(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("expected status 201, got %d", rr.Code)
	}

	requests := panel.live.requests.latest(1)
	if len(requests) != 1 {
		t.Fatalf("expected one request event, got %d", len(requests))
	}
	event := requests[0]
	if event.Status != http.StatusCreated {
		t.Fatalf("expected status=%d, got %d", http.StatusCreated, event.Status)
	}
	if event.TraceID != "trace-1" {
		t.Fatalf("expected trace_id=trace-1, got %q", event.TraceID)
	}
	if event.UserID != "user-42" {
		t.Fatalf("expected user_id=user-42, got %q", event.UserID)
	}
	if event.PayloadPreview == "" || event.PayloadPreview == "query:token=abc123&name=john" {
		t.Fatalf("expected redacted payload preview, got %q", event.PayloadPreview)
	}

	sessions := panel.live.sessions.snapshot(10)
	if len(sessions) != 1 {
		t.Fatalf("expected one tracked session, got %d", len(sessions))
	}
	if sessions[0].UserID != "user-42" {
		t.Fatalf("expected session user_id=user-42, got %q", sessions[0].UserID)
	}
}

func TestLiveTrafficMiddlewareSkipsWebSocketUpgrade(t *testing.T) {
	panel, cleanup := setupPanelForTest(t, db.EngineSQL)
	defer cleanup()

	mw := panel.liveTrafficMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusSwitchingProtocols)
	}))

	req := httptest.NewRequest(http.MethodGet, "/admin/api/live/ws", nil)
	req.Header.Set("Connection", "Upgrade")
	req.Header.Set("Upgrade", "websocket")

	rr := httptest.NewRecorder()
	mw.ServeHTTP(rr, req)

	if rr.Code != http.StatusSwitchingProtocols {
		t.Fatalf("expected status 101, got %d", rr.Code)
	}
	if got := panel.live.requests.latest(10); len(got) != 0 {
		t.Fatalf("expected no recorded events for websocket upgrade, got %d", len(got))
	}
}

func TestPanelLiveSnapshotEndpoint(t *testing.T) {
	panel, cleanup := setupPanelForTest(t, db.EngineSQL)
	defer cleanup()

	req := httptest.NewRequest(http.MethodGet, "/health?password=secret", nil)
	req = req.WithContext(observe.CtxWithRequestID(req.Context(), "req-2"))
	panel.recordLiveRequest(req, http.StatusOK, 12*time.Millisecond)
	panel.onModelSQLQuery(observe.CtxWithTraceID(context.Background(), "trace-2"), model.SQLQueryEvent{
		ModelName: "AdminUser",
		Operation: "select.list",
		Query:     "SELECT id, email FROM admin_users WHERE email = ? LIMIT ?",
		Args:      []interface{}{"admin@example.com", 25},
		Duration:  9 * time.Millisecond,
	})

	h := panel.Handler()
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/api/live/snapshot?limit=5", nil))

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rr.Code)
	}

	var payload liveSnapshotResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode response failed: %v body=%s", err, rr.Body.String())
	}
	if !payload.Enabled {
		t.Fatalf("expected live snapshot enabled")
	}
	if len(payload.Requests) == 0 {
		t.Fatalf("expected at least one request event")
	}
	if len(payload.Queries) == 0 {
		t.Fatalf("expected at least one sql query event")
	}
	if payload.SQLBuffer.Stored == 0 {
		t.Fatalf("expected sql buffer to store events")
	}
	if len(payload.Queries[0].Args) == 0 || payload.Queries[0].Args[0] != "string(17):***" {
		t.Fatalf("expected redacted string sql args, got %#v", payload.Queries[0].Args)
	}
	if payload.Stream.Published == 0 {
		t.Fatalf("expected published events > 0")
	}
}

func TestPanelOnModelSQLQueryPublishesEvent(t *testing.T) {
	panel, cleanup := setupPanelForTest(t, db.EngineSQL)
	defer cleanup()

	busCh, unsubscribe := panel.live.bus.subscribe()
	defer unsubscribe()

	ctx := observe.CtxWithRequestID(context.Background(), "req-sql-1")
	ctx = observe.CtxWithTraceID(ctx, "trace-sql-1")
	panel.onModelSQLQuery(ctx, model.SQLQueryEvent{
		ModelName: "AdminUser",
		Operation: "update",
		Query:     "UPDATE admin_users SET name = ? WHERE id = ?",
		Args:      []interface{}{"Alice", 7},
		Duration:  7 * time.Millisecond,
	})

	select {
	case event := <-busCh:
		if event.Type != "db.query" {
			t.Fatalf("expected db.query event type, got %q", event.Type)
		}
		if event.SQL == nil || event.SQL.TraceID != "trace-sql-1" {
			t.Fatalf("expected sql event with trace id, got %#v", event.SQL)
		}
		if len(event.SQL.Args) == 0 || event.SQL.Args[0] != "string(5):***" {
			t.Fatalf("expected first arg redacted, got %#v", event.SQL.Args)
		}
	case <-time.After(500 * time.Millisecond):
		t.Fatal("expected sql live event on bus")
	}
}

func TestHandleLiveWSRejectsInvalidOrigin(t *testing.T) {
	panel, cleanup := setupPanelForTest(t, db.EngineSQL)
	defer cleanup()

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/live/ws", nil)
	req.Header.Set("Origin", "https://evil.example")
	req.Host = "admin.example"
	panel.handleLiveWS(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Fatalf("expected status 403 for invalid origin, got %d", rr.Code)
	}
}
