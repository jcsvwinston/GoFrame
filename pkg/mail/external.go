package mail

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"time"
)

type externalSender struct {
	driver  string
	binary  string
	timeout time.Duration
}

func newExternalSender(driver, binary string, timeout time.Duration) Sender {
	if timeout <= 0 {
		timeout = 10 * time.Second
	}
	return &externalSender{
		driver:  driver,
		binary:  binary,
		timeout: timeout,
	}
}

func (s *externalSender) Send(ctx context.Context, msg Message) error {
	if err := validateMessage(msg); err != nil {
		return err
	}
	if ctx == nil {
		ctx = context.Background()
	}

	type pluginPayload struct {
		Driver  string            `json:"driver"`
		From    string            `json:"from"`
		To      []string          `json:"to"`
		Subject string            `json:"subject"`
		Body    string            `json:"body"`
		Headers map[string]string `json:"headers,omitempty"`
	}
	raw, err := json.Marshal(pluginPayload{
		Driver:  s.driver,
		From:    msg.From,
		To:      msg.To,
		Subject: msg.Subject,
		Body:    msg.Body,
		Headers: msg.Headers,
	})
	if err != nil {
		return fmt.Errorf("encode external mail payload: %w", err)
	}

	callCtx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	cmd := exec.CommandContext(callCtx, s.binary)
	cmd.Stdin = bytes.NewReader(raw)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		msg := stderr.String()
		if msg != "" {
			return fmt.Errorf("mail plugin %s failed: %w (%s)", s.binary, err, msg)
		}
		return fmt.Errorf("mail plugin %s failed: %w", s.binary, err)
	}
	return nil
}
