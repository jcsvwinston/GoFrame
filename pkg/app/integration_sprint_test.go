package app

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/jcsvwinston/nucleus/pkg/authz"
	"github.com/jcsvwinston/nucleus/pkg/circuit"
	"github.com/jcsvwinston/nucleus/pkg/mail"
	"github.com/jcsvwinston/nucleus/pkg/router"
)

// e2eFailingMailDriver is registered once at test package init so the
// E2E test below can configure App.New with a driver that always fails
// Send. Cross-package test cleanup is awkward (pkg/mail does not
// expose Unregister to other packages), so a one-shot registration
// with a stable name is the least invasive wiring.
const e2eFailingMailDriver = "e2e-adr004-failing"

var (
	e2eMailSendErr     = errors.New("e2e: simulated mail backend down")
	e2eMailSendCallsTo atomic.Int64
)

type e2eFailingSender struct{}

func (e2eFailingSender) Send(_ context.Context, _ mail.Message) error {
	e2eMailSendCallsTo.Add(1)
	return e2eMailSendErr
}

func init() {
	_ = mail.RegisterProvider(e2eFailingMailDriver, func(_ mail.Config) (mail.Sender, error) {
		return e2eFailingSender{}, nil
	})
}

// TestAppNew_ADR004IntegrationSprint_EndToEnd is the cross-integration
// test that the ADR-004 integration sprint deferred as its single open
// acceptance criterion. It builds one App.New with all three primitives
// active simultaneously — default-deny Casbin middleware, JWT manager
// from JWTKeys[] with an asymmetric key that auto-mounts JWKS, and a
// circuit breaker wrapping a mail provider that fails on Send — and
// exercises route paths that touch every concern:
//
//  1. /healthz and /.well-known/jwks.json pass the default-deny
//     middleware via the bootstrap allow-list.
//  2. An unauthenticated business route returns 403 (deny applies).
//  3. After AddPolicy(anonymous, …), the same route returns 200 (the
//     middleware respects allow rules).
//  4. The business route invokes App.Mailer.Send. The wrapped sender
//     returns the provider error on the first call; the second call
//     surfaces circuit.ErrOpen — proving the breaker is in the chain.
//  5. /healthz still returns 200 with the breaker open — Healthy()
//     bypasses the breaker as documented in pkg/mail/breaker.go.
//
// The test is silent on storage because the default storage provider
// is `local`, which is intentionally never wrapped. The storage half
// of the breaker autowrap is covered by pkg/storage/breaker_test.go.
func TestAppNew_ADR004IntegrationSprint_EndToEnd(t *testing.T) {
	e2eMailSendCallsTo.Store(0)

	priv := mustRSA(t)
	dir := t.TempDir()
	pemPath := filepath.Join(dir, "rsa-e2e.pem")
	if err := os.WriteFile(pemPath, mustEncodePKCS8(t, priv), 0o600); err != nil {
		t.Fatalf("write pem: %v", err)
	}

	cfg := testAppConfig()
	cfg.JWTExpiry = time.Hour
	cfg.JWTKeys = []JWTKeySpec{{KID: "rsa-e2e", Algorithm: "RS256", PemPath: pemPath}}
	cfg.JWTCurrentKID = "rsa-e2e"
	cfg.MailDriver = e2eFailingMailDriver
	cfg.MailCircuitBreaker = CircuitBreakerSpec{
		Enabled:               true,
		FailureThreshold:      1,
		Cooldown:              time.Hour, // long enough that the breaker stays open across calls
		HalfOpenMaxConcurrent: 1,
	}

	a, err := New(cfg)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	defer a.Shutdown(context.Background())

	// All three subsystems built.
	if a.JWT == nil || a.JWT.CurrentKID() != "rsa-e2e" {
		t.Fatalf("JWT not wired: %+v", a.JWT)
	}
	if a.Authorizer == nil {
		t.Fatal("Authorizer not constructed")
	}
	if a.Mailer == nil {
		t.Fatal("Mailer not constructed")
	}

	// --- 1. Bootstrap allow-list routes pass default-deny. ---

	for _, path := range []string{"/healthz", "/.well-known/jwks.json"} {
		req := httptest.NewRequest(http.MethodGet, path, nil)
		rec := httptest.NewRecorder()
		a.Router.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("bootstrap route %s: expected 200, got %d body=%s", path, rec.Code, rec.Body.String())
		}
	}

	// JWKS must contain the configured key.
	{
		req := httptest.NewRequest(http.MethodGet, "/.well-known/jwks.json", nil)
		rec := httptest.NewRecorder()
		a.Router.ServeHTTP(rec, req)
		body := rec.Body.String()
		if !strings.Contains(body, `"kid":"rsa-e2e"`) || !strings.Contains(body, `"kty":"RSA"`) {
			t.Fatalf("JWKS body missing expected fields: %s", body)
		}
	}

	// --- 2. Business route: 403 under default-deny. ---

	a.Router.Get("/api/send-mail", func(c *router.Context) error {
		err := a.Mailer.Send(c.Request.Context(), mail.Message{
			From:    "from@example.com",
			To:      []string{"to@example.com"},
			Subject: "e2e",
			Body:    "hello",
		})
		if err != nil {
			return c.JSON(http.StatusBadGateway, map[string]string{"error": err.Error()})
		}
		return c.JSON(http.StatusOK, map[string]string{"sent": "true"})
	})

	{
		req := httptest.NewRequest(http.MethodGet, "/api/send-mail", nil)
		rec := httptest.NewRecorder()
		a.Router.ServeHTTP(rec, req)
		if rec.Code != http.StatusForbidden {
			t.Fatalf("expected 403 on /api/send-mail under default-deny, got %d body=%s", rec.Code, rec.Body.String())
		}
	}

	// --- 3. Anonymous allow opens the route. ---

	if err := a.Authorizer.AddPolicy(authz.BootstrapSubject, "/api/send-mail", "*"); err != nil {
		t.Fatalf("AddPolicy: %v", err)
	}

	// --- 4. First call: mail provider error surfaces. Second call: breaker open. ---

	callBody := func() (int, map[string]string) {
		req := httptest.NewRequest(http.MethodGet, "/api/send-mail", nil)
		rec := httptest.NewRecorder()
		a.Router.ServeHTTP(rec, req)
		var body map[string]string
		_ = json.Unmarshal(rec.Body.Bytes(), &body)
		return rec.Code, body
	}

	code, body := callBody()
	if code != http.StatusBadGateway {
		t.Fatalf("call 1: expected 502 from mail error, got %d body=%v", code, body)
	}
	if !strings.Contains(body["error"], e2eMailSendErr.Error()) {
		t.Fatalf("call 1: expected provider error surfaced, got %v", body)
	}

	code, body = callBody()
	if code != http.StatusBadGateway {
		t.Fatalf("call 2: expected 502 with ErrOpen, got %d body=%v", code, body)
	}
	if !strings.Contains(body["error"], circuit.ErrOpen.Error()) {
		t.Fatalf("call 2: expected circuit.ErrOpen, got %v", body)
	}

	// --- 5. /healthz unaffected by breaker state. ---
	{
		req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
		rec := httptest.NewRecorder()
		a.Router.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("/healthz after breaker open: expected 200, got %d body=%s", rec.Code, rec.Body.String())
		}
		// The body should still report healthy overall — the breaker is a
		// runtime decoration of Send, not of the health probe.
		body := rec.Body.String()
		if !strings.Contains(body, `"status":"healthy"`) && !strings.Contains(body, `"status":"ok"`) {
			t.Fatalf("/healthz body unexpectedly degraded: %s", body)
		}
	}

	// Only the first send reached the underlying provider; the second
	// short-circuited at the breaker. This is the load-bearing proof
	// that the breaker is actually in the chain — without it, two
	// inner Send calls would have been recorded.
	if got := e2eMailSendCallsTo.Load(); got != 1 {
		t.Fatalf("expected exactly 1 inner Send (call 2 should short-circuit), got %d", got)
	}
}
