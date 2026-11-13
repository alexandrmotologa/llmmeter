package alerts

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"
)

func TestDispatchWebhook(t *testing.T) {
	var received WebhookPayload
	var mu sync.Mutex
	receivedChan := make(chan bool, 1)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		defer mu.Unlock()
		_ = json.NewDecoder(r.Body).Decode(&received)
		w.WriteHeader(http.StatusOK)
		receivedChan <- true
	}))
	defer server.Close()

	payload := WebhookPayload{
		Event:     "budget_exceeded",
		TenantID:  "tenant-test",
		Type:      "budget",
		Severity:  "critical",
		Message:   "Tenant exceeded budget limit",
		SpendUSD:  120.50,
		BudgetUSD: 100.00,
		Timestamp: time.Now(),
	}

	Dispatch(server.URL, payload)

	select {
	case <-receivedChan:
		mu.Lock()
		defer mu.Unlock()
		if received.TenantID != "tenant-test" || received.Event != "budget_exceeded" {
			t.Errorf("unexpected payload received: %+v", received)
		}
	case <-time.After(2 * time.Second):
		t.Errorf("timeout waiting for webhook delivery")
	}
}
