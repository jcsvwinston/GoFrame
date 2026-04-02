package mail

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/smtp"
	"strconv"
	"strings"
	"time"
)

type smtpSender struct {
	host    string
	port    int
	user    string
	pass    string
	timeout time.Duration
}

func newSMTPSender(cfg Config) (Sender, error) {
	host := strings.TrimSpace(cfg.SMTPHost)
	if host == "" {
		return nil, fmt.Errorf("smtp host is required for smtp driver")
	}
	if cfg.SMTPPort <= 0 {
		return nil, fmt.Errorf("smtp port must be greater than 0 for smtp driver")
	}
	timeout := cfg.Timeout
	if timeout <= 0 {
		timeout = 10 * time.Second
	}
	return &smtpSender{
		host:    host,
		port:    cfg.SMTPPort,
		user:    strings.TrimSpace(cfg.SMTPUser),
		pass:    cfg.SMTPPass,
		timeout: timeout,
	}, nil
}

func (s *smtpSender) Send(ctx context.Context, msg Message) error {
	if err := validateMessage(msg); err != nil {
		return err
	}
	if ctx == nil {
		ctx = context.Background()
	}

	addr := net.JoinHostPort(s.host, strconv.Itoa(s.port))
	dialer := &net.Dialer{Timeout: s.timeout}
	conn, err := dialer.DialContext(ctx, "tcp", addr)
	if err != nil {
		return fmt.Errorf("connect smtp server %s: %w", addr, err)
	}
	_ = conn.SetDeadline(time.Now().Add(s.timeout))

	client, err := smtp.NewClient(conn, s.host)
	if err != nil {
		_ = conn.Close()
		return fmt.Errorf("start smtp client: %w", err)
	}
	defer client.Close()

	if ok, _ := client.Extension("STARTTLS"); ok {
		if err := client.StartTLS(&tls.Config{ServerName: s.host}); err != nil {
			return fmt.Errorf("smtp starttls failed: %w", err)
		}
	}

	if s.user != "" {
		auth := smtp.PlainAuth("", s.user, s.pass, s.host)
		if err := client.Auth(auth); err != nil {
			return fmt.Errorf("smtp auth failed: %w", err)
		}
	}

	from := strings.TrimSpace(msg.From)
	if err := client.Mail(from); err != nil {
		return fmt.Errorf("smtp MAIL FROM failed: %w", err)
	}
	for _, recipient := range cleanRecipients(msg.To) {
		if err := client.Rcpt(recipient); err != nil {
			return fmt.Errorf("smtp RCPT TO failed for %s: %w", recipient, err)
		}
	}

	wc, err := client.Data()
	if err != nil {
		return fmt.Errorf("smtp DATA failed: %w", err)
	}
	payload := buildRFC822Message(msg)
	if _, err := wc.Write(payload); err != nil {
		_ = wc.Close()
		return fmt.Errorf("write smtp payload: %w", err)
	}
	if err := wc.Close(); err != nil {
		return fmt.Errorf("finalize smtp payload: %w", err)
	}

	if err := client.Quit(); err != nil {
		return fmt.Errorf("smtp quit failed: %w", err)
	}
	return nil
}
