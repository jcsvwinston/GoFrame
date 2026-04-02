package mail

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const defaultSendGridEndpoint = "https://api.sendgrid.com/v3/mail/send"

type sendGridSender struct {
	apiKey   string
	endpoint string
	client   *http.Client
}

func newSendGridSender(cfg Config) (Sender, error) {
	apiKey := strings.TrimSpace(cfg.SendGridAPIKey)
	if apiKey == "" {
		return nil, fmt.Errorf("sendgrid api key is required for sendgrid driver")
	}
	endpoint := strings.TrimSpace(cfg.SendGridEndpoint)
	if endpoint == "" {
		endpoint = defaultSendGridEndpoint
	}
	timeout := cfg.Timeout
	if timeout <= 0 {
		timeout = 10 * time.Second
	}
	return &sendGridSender{
		apiKey:   apiKey,
		endpoint: endpoint,
		client:   &http.Client{Timeout: timeout},
	}, nil
}

func (s *sendGridSender) Send(ctx context.Context, msg Message) error {
	if err := validateMessage(msg); err != nil {
		return err
	}
	if ctx == nil {
		ctx = context.Background()
	}

	type recipient struct {
		Email string `json:"email"`
	}
	type payload struct {
		Personalizations []struct {
			To []recipient `json:"to"`
		} `json:"personalizations"`
		From struct {
			Email string `json:"email"`
		} `json:"from"`
		Subject string `json:"subject"`
		Content []struct {
			Type  string `json:"type"`
			Value string `json:"value"`
		} `json:"content"`
	}

	toList := make([]recipient, 0, len(msg.To))
	for _, r := range cleanRecipients(msg.To) {
		toList = append(toList, recipient{Email: r})
	}

	var body payload
	body.Personalizations = []struct {
		To []recipient `json:"to"`
	}{{To: toList}}
	body.From.Email = strings.TrimSpace(msg.From)
	body.Subject = strings.TrimSpace(msg.Subject)
	body.Content = []struct {
		Type  string `json:"type"`
		Value string `json:"value"`
	}{{Type: "text/plain", Value: msg.Body}}

	raw, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("encode sendgrid payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.endpoint, bytes.NewReader(raw))
	if err != nil {
		return fmt.Errorf("build sendgrid request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+s.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("sendgrid request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return nil
	}

	respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 2048))
	return fmt.Errorf("sendgrid returned status %d: %s", resp.StatusCode, strings.TrimSpace(string(respBody)))
}
