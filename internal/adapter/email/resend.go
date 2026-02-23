package email

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/life-journaling/core/internal/domain"
)

const resendAPIURL = "https://api.resend.com/emails"

// ResendProvider implements IEmailProvider using the Resend API.
type ResendProvider struct {
	apiKey    string
	fromEmail string
	client    *http.Client
}

// NewResendProvider creates a new ResendProvider.
func NewResendProvider(apiKey string, fromEmail string) *ResendProvider {
	return &ResendProvider{
		apiKey:    apiKey,
		fromEmail: fromEmail,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// resendRequest represents the Resend API email payload.
type resendRequest struct {
	From    string `json:"from"`
	To      string `json:"to"`
	Subject string `json:"subject"`
	Text    string `json:"text"`
}

// SendPrompt sends a prompt email via the Resend API.
func (p *ResendProvider) SendPrompt(ctx context.Context, toEmail string, subject string, body string) error {
	payload := resendRequest{
		From:    p.fromEmail,
		To:      toEmail,
		Subject: subject,
		Text:    body,
	}

	jsonBody, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshaling email payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, resendAPIURL, bytes.NewReader(jsonBody))
	if err != nil {
		return fmt.Errorf("creating request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+p.apiKey)

	resp, err := p.client.Do(req)
	if err != nil {
		return fmt.Errorf("%w: %v", domain.ErrEmailSendFailed, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("%w: status %d, body: %s", domain.ErrEmailSendFailed, resp.StatusCode, string(respBody))
	}

	return nil
}
