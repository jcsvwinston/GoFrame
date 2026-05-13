package health

import (
	"context"
	"errors"

	"github.com/jcsvwinston/nucleus/pkg/mail"
)

// NewMailProbe builds a Prober that exercises a mail.Sender's
// optional health check. The sender must implement mail.HealthChecker;
// callers should type-assert before calling NewMailProbe, and skip
// registration if the assertion fails — there is intentionally no
// silent fallback so the /healthz response never includes a
// information-free "skipped" row.
//
// SupportsMailProbe is exposed as a convenience wrapper around the
// type-assertion check so callers can keep the conditional outside
// this package.
func NewMailProbe(name string, sender mail.Sender) Prober {
	return &mailProbe{name: name, sender: sender}
}

// SupportsMailProbe reports whether the given Sender implements the
// optional mail.HealthChecker contract.
func SupportsMailProbe(sender mail.Sender) bool {
	if sender == nil {
		return false
	}
	_, ok := sender.(mail.HealthChecker)
	return ok
}

type mailProbe struct {
	name   string
	sender mail.Sender
}

func (p *mailProbe) Name() string { return p.name }

func (p *mailProbe) Probe(ctx context.Context) error {
	if p.sender == nil {
		return errors.New("mail sender is nil")
	}
	hc, ok := p.sender.(mail.HealthChecker)
	if !ok {
		return errors.New("mail sender does not implement HealthChecker")
	}
	return hc.Healthy(ctx)
}
