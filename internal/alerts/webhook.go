package alerts

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"time"
)

// WebhookPayload represents the standard alert notification sent to external services.
type WebhookPayload struct {
	Event     string    `json:"event"`
	TenantID  string    `json:"tenant_id"`
	Type      string    `json:"type"`
	Severity  string    `json:"severity"`
	Message   string    `json:"message"`
	SpendUSD  float64   `json:"spend_usd,omitempty"`
	BudgetUSD float64   `json:"budget_usd,omitempty"`
	Timestamp time.Time `json:"timestamp"`
}

// Dispatch sends a webhook notification asynchronously.
func Dispatch(webhookURL string, payload WebhookPayload) {
	if webhookURL == "" {
		return
	}

	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		data, err := json.Marshal(payload)
		if err != nil {
			return
		}

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, webhookURL, bytes.NewReader(data))
		if err != nil {
			return
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("User-Agent", "llmmeter-finops/1.0")

		client := &http.Client{Timeout: 5 * time.Second}
		resp, err := client.Do(req)
		if err == nil {
			_ = resp.Body.Close()
		}
	}()
}
