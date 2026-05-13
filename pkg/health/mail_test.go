package health

import (
	"context"
	"errors"
	"testing"

	"github.com/jcsvwinston/nucleus/pkg/mail"
)

type stubSenderWithHealth struct {
	healthErr error
}

func (s *stubSenderWithHealth) Send(context.Context, mail.Message) error { return nil }
func (s *stubSenderWithHealth) Healthy(context.Context) error            { return s.healthErr }

type stubSenderNoHealth struct{}

func (s *stubSenderNoHealth) Send(context.Context, mail.Message) error { return nil }

func TestSupportsMailProbe(t *testing.T) {
	if !SupportsMailProbe(&stubSenderWithHealth{}) {
		t.Fatal("sender implementing HealthChecker should be probable")
	}
	if SupportsMailProbe(&stubSenderNoHealth{}) {
		t.Fatal("sender without HealthChecker must not be probable")
	}
	if SupportsMailProbe(nil) {
		t.Fatal("nil sender must not be probable")
	}
}

func TestMailProbe_HealthyAndUnhealthy(t *testing.T) {
	p := NewMailProbe("mail", &stubSenderWithHealth{})
	if err := p.Probe(context.Background()); err != nil {
		t.Fatalf("expected healthy, got %v", err)
	}
	if p.Name() != "mail" {
		t.Fatalf("unexpected name: %s", p.Name())
	}

	bad := NewMailProbe("mail", &stubSenderWithHealth{healthErr: errors.New("connection refused")})
	if err := bad.Probe(context.Background()); err == nil {
		t.Fatal("expected error from unhealthy sender")
	}
}

func TestMailProbe_NilSenderSurfacesError(t *testing.T) {
	p := NewMailProbe("mail", nil)
	if err := p.Probe(context.Background()); err == nil {
		t.Fatal("expected error for nil sender")
	}
}

func TestMailProbe_SenderWithoutHealthCheckerSurfacesError(t *testing.T) {
	// Constructed via NewMailProbe even though SupportsMailProbe returns
	// false — Probe must still degrade loudly so the operator notices a
	// misconfigured probe rather than silently passing it.
	p := NewMailProbe("mail", &stubSenderNoHealth{})
	if err := p.Probe(context.Background()); err == nil {
		t.Fatal("expected error when sender lacks HealthChecker")
	}
}
