package admin

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/jcsvwinston/GoFrame/pkg/auth"
	"github.com/jcsvwinston/GoFrame/pkg/observe"
	"github.com/jcsvwinston/GoFrame/pkg/router"
	"golang.org/x/net/websocket"
)

const (
	defaultLiveRequestBufferSize = 256
	defaultLiveSubscriberBuffer  = 128
	defaultLiveSessionTTL        = 30 * time.Minute
	defaultLiveListLimit         = 50
	maxLiveListLimit             = 1000
)

type liveTrafficObservedKey struct{}

type liveRuntime struct {
	requests *requestRingBuffer
	bus      *liveEventBus
	sessions *liveSessionStore
}

type liveSnapshotResponse struct {
	Enabled       bool                   `json:"enabled"`
	GeneratedAt   string                 `json:"generated_at"`
	Limit         int                    `json:"limit"`
	Requests      []liveRequestEvent     `json:"requests"`
	Sessions      []liveSessionActivity  `json:"sessions"`
	Stream        liveStreamStats        `json:"stream"`
	RequestBuffer liveRequestBufferStats `json:"request_buffer"`
}

type liveRequestBufferStats struct {
	Capacity int `json:"capacity"`
	Stored   int `json:"stored"`
}

type liveStreamStats struct {
	Subscribers int    `json:"subscribers"`
	Published   uint64 `json:"published"`
	Dropped     uint64 `json:"dropped"`
}

type liveRequestEvent struct {
	Timestamp      string `json:"timestamp"`
	Method         string `json:"method"`
	Path           string `json:"path"`
	Status         int    `json:"status"`
	DurationMS     int64  `json:"duration_ms"`
	RequestID      string `json:"request_id,omitempty"`
	TraceID        string `json:"trace_id,omitempty"`
	UserID         string `json:"user_id,omitempty"`
	RemoteIP       string `json:"remote_ip,omitempty"`
	UserAgent      string `json:"user_agent,omitempty"`
	PayloadPreview string `json:"payload_preview,omitempty"`
}

type liveSessionActivity struct {
	SessionToken string `json:"session_token,omitempty"`
	TokenShort   string `json:"token_short"`
	UserID       string `json:"user_id,omitempty"`
	IP           string `json:"ip,omitempty"`
	UserAgent    string `json:"user_agent,omitempty"`
	LastRoute    string `json:"last_route"`
	LastSeenAt   string `json:"last_seen_at"`
	TraceID      string `json:"trace_id,omitempty"`
}

type liveEventEnvelope struct {
	Type      string               `json:"type"`
	Timestamp string               `json:"timestamp"`
	Request   *liveRequestEvent    `json:"request,omitempty"`
	Session   *liveSessionActivity `json:"session,omitempty"`
}

type requestRingBuffer struct {
	mu       sync.RWMutex
	events   []liveRequestEvent
	head     int
	size     int
	capacity int
}

type liveEventBus struct {
	mu             sync.RWMutex
	nextID         uint64
	subscriberSize int
	subscribers    map[uint64]chan liveEventEnvelope
	published      atomic.Uint64
	dropped        atomic.Uint64
}

type liveSessionStore struct {
	mu      sync.RWMutex
	entries map[string]liveSessionActivity
	ttl     time.Duration
}

func newLiveRuntime() *liveRuntime {
	return &liveRuntime{
		requests: newRequestRingBuffer(defaultLiveRequestBufferSize),
		bus:      newLiveEventBus(defaultLiveSubscriberBuffer),
		sessions: newLiveSessionStore(defaultLiveSessionTTL),
	}
}

func newRequestRingBuffer(capacity int) *requestRingBuffer {
	if capacity <= 0 {
		capacity = defaultLiveRequestBufferSize
	}
	return &requestRingBuffer{
		events:   make([]liveRequestEvent, capacity),
		capacity: capacity,
	}
}

func (rb *requestRingBuffer) push(event liveRequestEvent) {
	rb.mu.Lock()
	defer rb.mu.Unlock()

	rb.events[rb.head] = event
	rb.head = (rb.head + 1) % rb.capacity
	if rb.size < rb.capacity {
		rb.size++
	}
}

func (rb *requestRingBuffer) latest(limit int) []liveRequestEvent {
	rb.mu.RLock()
	defer rb.mu.RUnlock()

	if rb.size == 0 || limit <= 0 {
		return []liveRequestEvent{}
	}
	if limit > rb.size {
		limit = rb.size
	}

	out := make([]liveRequestEvent, 0, limit)
	for i := 0; i < limit; i++ {
		idx := (rb.head - 1 - i + rb.capacity) % rb.capacity
		out = append(out, rb.events[idx])
	}
	return out
}

func (rb *requestRingBuffer) stats() liveRequestBufferStats {
	rb.mu.RLock()
	defer rb.mu.RUnlock()
	return liveRequestBufferStats{
		Capacity: rb.capacity,
		Stored:   rb.size,
	}
}

func newLiveEventBus(subscriberSize int) *liveEventBus {
	if subscriberSize <= 0 {
		subscriberSize = defaultLiveSubscriberBuffer
	}
	return &liveEventBus{
		subscriberSize: subscriberSize,
		subscribers:    make(map[uint64]chan liveEventEnvelope),
	}
}

func (b *liveEventBus) subscribe() (<-chan liveEventEnvelope, func()) {
	id := atomic.AddUint64(&b.nextID, 1)
	ch := make(chan liveEventEnvelope, b.subscriberSize)

	b.mu.Lock()
	b.subscribers[id] = ch
	b.mu.Unlock()

	unsubscribe := func() {
		b.mu.Lock()
		existing, ok := b.subscribers[id]
		if ok {
			delete(b.subscribers, id)
			close(existing)
		}
		b.mu.Unlock()
	}
	return ch, unsubscribe
}

func (b *liveEventBus) publish(event liveEventEnvelope) {
	if b == nil {
		return
	}
	b.published.Add(1)

	b.mu.RLock()
	defer b.mu.RUnlock()

	for _, ch := range b.subscribers {
		select {
		case ch <- event:
		default:
			b.dropped.Add(1)
		}
	}
}

func (b *liveEventBus) stats() liveStreamStats {
	if b == nil {
		return liveStreamStats{}
	}
	b.mu.RLock()
	subs := len(b.subscribers)
	b.mu.RUnlock()
	return liveStreamStats{
		Subscribers: subs,
		Published:   b.published.Load(),
		Dropped:     b.dropped.Load(),
	}
}

func newLiveSessionStore(ttl time.Duration) *liveSessionStore {
	if ttl <= 0 {
		ttl = defaultLiveSessionTTL
	}
	return &liveSessionStore{
		entries: make(map[string]liveSessionActivity),
		ttl:     ttl,
	}
}

func (s *liveSessionStore) upsert(key string, value liveSessionActivity) {
	if s == nil || strings.TrimSpace(key) == "" {
		return
	}
	s.mu.Lock()
	s.entries[key] = value
	s.gcLocked(time.Now().UTC())
	s.mu.Unlock()
}

func (s *liveSessionStore) snapshot(limit int) []liveSessionActivity {
	if s == nil || limit <= 0 {
		return []liveSessionActivity{}
	}

	now := time.Now().UTC()
	s.mu.Lock()
	s.gcLocked(now)

	rows := make([]liveSessionActivity, 0, len(s.entries))
	for _, row := range s.entries {
		rows = append(rows, row)
	}
	s.mu.Unlock()

	sort.SliceStable(rows, func(i, j int) bool {
		ti := parseRFC3339(rows[i].LastSeenAt)
		tj := parseRFC3339(rows[j].LastSeenAt)
		if !ti.Equal(tj) {
			return ti.After(tj)
		}
		return rows[i].TokenShort < rows[j].TokenShort
	})

	if len(rows) > limit {
		rows = rows[:limit]
	}
	return rows
}

func (s *liveSessionStore) gcLocked(now time.Time) {
	if s == nil {
		return
	}
	for key, row := range s.entries {
		seenAt := parseRFC3339(row.LastSeenAt)
		if seenAt.IsZero() || now.Sub(seenAt) <= s.ttl {
			continue
		}
		delete(s.entries, key)
	}
}

func parseRFC3339(raw string) time.Time {
	value := strings.TrimSpace(raw)
	if value == "" {
		return time.Time{}
	}
	ts, err := time.Parse(time.RFC3339, value)
	if err != nil {
		return time.Time{}
	}
	return ts
}

func (p *Panel) liveTrafficMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if p == nil || p.live == nil || r == nil {
			next.ServeHTTP(w, r)
			return
		}
		if r.Context().Value(liveTrafficObservedKey{}) != nil {
			next.ServeHTTP(w, r)
			return
		}
		if isWebSocketUpgrade(r) {
			next.ServeHTTP(w, r.WithContext(contextWithLiveObserved(r)))
			return
		}

		start := time.Now()
		ww := router.NewWrapResponseWriter(w, r.ProtoMajor)
		next.ServeHTTP(ww, r.WithContext(contextWithLiveObserved(r)))
		p.recordLiveRequest(r, ww.Status(), time.Since(start))
	})
}

func contextWithLiveObserved(r *http.Request) context.Context {
	if r == nil {
		return nil
	}
	return context.WithValue(r.Context(), liveTrafficObservedKey{}, true)
}

func isWebSocketUpgrade(r *http.Request) bool {
	if r == nil {
		return false
	}
	connection := strings.ToLower(strings.TrimSpace(r.Header.Get("Connection")))
	upgrade := strings.ToLower(strings.TrimSpace(r.Header.Get("Upgrade")))
	return strings.Contains(connection, "upgrade") && upgrade == "websocket"
}

func (p *Panel) recordLiveRequest(r *http.Request, status int, duration time.Duration) {
	if p == nil || p.live == nil || r == nil {
		return
	}

	now := time.Now().UTC()
	ctx := r.Context()
	event := liveRequestEvent{
		Timestamp:      now.Format(time.RFC3339),
		Method:         r.Method,
		Path:           truncateText(r.URL.Path, 240),
		Status:         status,
		DurationMS:     duration.Milliseconds(),
		RequestID:      strings.TrimSpace(observe.RequestIDFromCtx(ctx)),
		TraceID:        strings.TrimSpace(observe.TraceIDFromCtx(ctx)),
		UserID:         strings.TrimSpace(observe.UserIDFromCtx(ctx)),
		RemoteIP:       auth.ClientIPFromRequest(r),
		UserAgent:      truncateText(strings.TrimSpace(r.UserAgent()), 320),
		PayloadPreview: livePayloadPreview(r),
	}
	p.live.requests.push(event)
	p.live.bus.publish(liveEventEnvelope{
		Type:      "http.request",
		Timestamp: event.Timestamp,
		Request:   &event,
	})

	p.recordLiveSessionActivity(r, now, event.TraceID)
}

func (p *Panel) recordLiveSessionActivity(r *http.Request, now time.Time, traceID string) {
	if p == nil || p.live == nil || r == nil {
		return
	}

	key, token := liveSessionKey(p.config.Session, r.Context())
	if key == "" {
		return
	}

	activity := liveSessionActivity{
		SessionToken: token,
		TokenShort:   shortenToken(token),
		UserID:       strings.TrimSpace(observe.UserIDFromCtx(r.Context())),
		IP:           auth.ClientIPFromRequest(r),
		UserAgent:    truncateText(strings.TrimSpace(r.UserAgent()), 320),
		LastRoute:    truncateText(strings.TrimSpace(r.URL.Path), 240),
		LastSeenAt:   now.Format(time.RFC3339),
		TraceID:      strings.TrimSpace(traceID),
	}
	if activity.UserID == "" && p.config.Auth != nil {
		if user, _ := p.authenticatedUser(r); user != nil {
			activity.UserID = strings.TrimSpace(user.ID)
		}
	}
	if activity.TokenShort == "" {
		activity.TokenShort = shortenToken(key)
	}

	p.live.sessions.upsert(key, activity)
	p.live.bus.publish(liveEventEnvelope{
		Type:      "session.activity",
		Timestamp: activity.LastSeenAt,
		Session:   &activity,
	})
}

func liveSessionKey(sm *auth.SessionManager, ctx context.Context) (key string, token string) {
	if sm != nil && sessionContextReady(sm, ctx) {
		token = strings.TrimSpace(sm.SCS().Token(ctx))
		if token != "" {
			return "session:" + token, token
		}
	}
	reqID := strings.TrimSpace(observe.RequestIDFromCtx(ctx))
	if reqID != "" {
		return "request:" + reqID, reqID
	}
	return "", ""
}

func livePayloadPreview(r *http.Request) string {
	if r == nil {
		return ""
	}

	method := strings.ToUpper(strings.TrimSpace(r.Method))
	switch method {
	case http.MethodGet, http.MethodDelete:
		q := redactSensitiveQuery(r.URL.Query())
		encoded := q.Encode()
		if encoded == "" {
			return ""
		}
		if len(encoded) > 180 {
			encoded = encoded[:180] + "..."
		}
		return truncateText("query:"+encoded, 220)
	default:
		if r.ContentLength > 0 {
			return truncateText(fmt.Sprintf("body:redacted (%d bytes)", r.ContentLength), 220)
		}
		ct := strings.ToLower(strings.TrimSpace(r.Header.Get("Content-Type")))
		if ct != "" {
			return truncateText("body:redacted ("+ct+")", 220)
		}
		return "body:redacted"
	}
}

func redactSensitiveQuery(values url.Values) url.Values {
	out := url.Values{}
	for key, items := range values {
		sensitive := isSensitiveKey(key)
		for _, item := range items {
			if sensitive {
				out.Add(key, "***")
				continue
			}
			out.Add(key, item)
		}
	}
	return out
}

func isSensitiveKey(key string) bool {
	normalized := strings.ToUpper(strings.TrimSpace(key))
	return strings.Contains(normalized, "KEY") ||
		strings.Contains(normalized, "SECRET") ||
		strings.Contains(normalized, "PASSWORD") ||
		strings.Contains(normalized, "TOKEN")
}

func truncateText(value string, maxLen int) string {
	text := strings.TrimSpace(value)
	if maxLen <= 0 || len(text) <= maxLen {
		return text
	}
	if maxLen <= 3 {
		return text[:maxLen]
	}
	return text[:maxLen-3] + "..."
}

func parseLiveListLimit(r *http.Request, fallback int) int {
	if fallback <= 0 {
		fallback = defaultLiveListLimit
	}
	if r == nil {
		return fallback
	}
	raw := strings.TrimSpace(r.URL.Query().Get("limit"))
	if raw == "" {
		return fallback
	}
	value, err := strconv.Atoi(raw)
	if err != nil || value <= 0 {
		return fallback
	}
	if value > maxLiveListLimit {
		return maxLiveListLimit
	}
	return value
}

func (p *Panel) handleLiveSnapshot(w http.ResponseWriter, r *http.Request) {
	if !p.authorizeAction(w, r, "*", "live_traffic") {
		return
	}
	now := time.Now().UTC()
	limit := parseLiveListLimit(r, defaultLiveListLimit)

	resp := liveSnapshotResponse{
		Enabled:     p != nil && p.live != nil,
		GeneratedAt: now.Format(time.RFC3339),
		Limit:       limit,
		Requests:    []liveRequestEvent{},
		Sessions:    []liveSessionActivity{},
	}
	if p == nil || p.live == nil {
		writeJSON(w, http.StatusOK, resp)
		return
	}

	resp.Requests = p.live.requests.latest(limit)
	resp.Sessions = p.live.sessions.snapshot(limit)
	resp.Stream = p.live.bus.stats()
	resp.RequestBuffer = p.live.requests.stats()
	writeJSON(w, http.StatusOK, resp)
}

func (p *Panel) handleLiveWS(w http.ResponseWriter, r *http.Request) {
	if !p.authorizeAction(w, r, "*", "live_traffic") {
		return
	}
	if p == nil || p.live == nil {
		writeErr(w, fmt.Errorf("live runtime is not enabled"))
		return
	}
	if !allowLiveWSOrigin(r) {
		http.Error(w, "websocket origin not allowed", http.StatusForbidden)
		return
	}

	websocket.Handler(func(conn *websocket.Conn) {
		defer conn.Close()
		conn.PayloadType = websocket.TextFrame

		ch, unsubscribe := p.live.bus.subscribe()
		defer unsubscribe()

		hello := liveEventEnvelope{
			Type:      "stream.ready",
			Timestamp: time.Now().UTC().Format(time.RFC3339),
		}
		if err := websocket.JSON.Send(conn, hello); err != nil {
			return
		}

		for event := range ch {
			if err := websocket.JSON.Send(conn, event); err != nil {
				return
			}
		}
	}).ServeHTTP(w, r)
}

func allowLiveWSOrigin(r *http.Request) bool {
	if r == nil {
		return false
	}
	originRaw := strings.TrimSpace(r.Header.Get("Origin"))
	if originRaw == "" {
		return true
	}
	origin, err := url.Parse(originRaw)
	if err != nil {
		return false
	}
	return strings.EqualFold(origin.Host, r.Host)
}
