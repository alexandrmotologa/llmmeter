package metrics

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestMetricsCollector(t *testing.T) {
	c := Init(nil)

	c.RecordRequest("tenant-a", "openai", "gpt-4o", 200, 100, 50, 20, 0.001)
	c.RecordCacheHit()
	c.RecordCacheMiss()
	c.RecordCircuitBreakerTrip()

	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	w := httptest.NewRecorder()

	c.Handler()(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", w.Code)
	}

	body := w.Body.String()
	if !strings.Contains(body, "llmmeter_http_requests_total") {
		t.Errorf("missing llmmeter_http_requests_total in output: %s", body)
	}
	if !strings.Contains(body, `tenant="tenant-a"`) {
		t.Errorf("missing tenant-a label in metrics: %s", body)
	}
	if !strings.Contains(body, "llmmeter_cache_hits_total 1") {
		t.Errorf("missing cache hits in metrics: %s", body)
	}
}
