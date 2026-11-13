package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/alexandrmotologa/llmmeter/internal/metrics"
	"github.com/alexandrmotologa/llmmeter/internal/store"
)

func TestAPIServer(t *testing.T) {
	tmpDir, _ := os.MkdirTemp("", "llmmeter-api-test-*")
	defer os.RemoveAll(tmpDir)

	s, _ := store.New(filepath.Join(tmpDir, "finops.db"))
	defer s.Close()

	metrics.Init(s)

	_ = s.CreateTenant(store.Tenant{
		ID:             "tenant-delta",
		Name:           "Delta Corp",
		MonthlyPlanUSD: 99.0,
	})
	_ = s.LogRequest(store.RequestLog{
		TenantID:         "tenant-delta",
		FeatureTag:       "ai-support",
		Provider:         "openai",
		Model:            "gpt-4o",
		PromptTokens:     2000,
		CompletionTokens: 500,
		CachedTokens:     800,
		CostUSD:          0.01,
		LatencyMS:        250,
		Timestamp:        time.Now(),
	})

	server := NewServer(s)
	mux := http.NewServeMux()
	server.RegisterRoutes(mux)

	// 1. GET /api/overview
	reqOverview := httptest.NewRequest(http.MethodGet, "/api/overview", nil)
	wOverview := httptest.NewRecorder()
	mux.ServeHTTP(wOverview, reqOverview)

	if wOverview.Code != http.StatusOK {
		t.Fatalf("expected 200 OK for overview, got %d", wOverview.Code)
	}

	// 2. GET /api/cache-analytics
	reqCache := httptest.NewRequest(http.MethodGet, "/api/cache-analytics", nil)
	wCache := httptest.NewRecorder()
	mux.ServeHTTP(wCache, reqCache)

	if wCache.Code != http.StatusOK {
		t.Fatalf("expected 200 OK for cache analytics, got %d", wCache.Code)
	}
	var ca store.CacheAnalytics
	_ = json.Unmarshal(wCache.Body.Bytes(), &ca)
	if ca.TotalCachedTokens != 800 {
		t.Errorf("expected 800 cached tokens, got %d", ca.TotalCachedTokens)
	}

	// 3. GET /api/tenants/tenant-delta/statement (JSON)
	reqStmt := httptest.NewRequest(http.MethodGet, "/api/tenants/tenant-delta/statement", nil)
	wStmt := httptest.NewRecorder()
	mux.ServeHTTP(wStmt, reqStmt)

	if wStmt.Code != http.StatusOK {
		t.Fatalf("expected 200 OK for statement, got %d", wStmt.Code)
	}
	var stmt store.TenantStatement
	_ = json.Unmarshal(wStmt.Body.Bytes(), &stmt)
	if stmt.TenantName != "Delta Corp" || len(stmt.LineItems) != 1 {
		t.Errorf("unexpected statement data: %+v", stmt)
	}

	// 4. GET /api/tenants/tenant-delta/statement?format=html
	reqStmtHTML := httptest.NewRequest(http.MethodGet, "/api/tenants/tenant-delta/statement?format=html", nil)
	wStmtHTML := httptest.NewRecorder()
	mux.ServeHTTP(wStmtHTML, reqStmtHTML)

	if wStmtHTML.Code != http.StatusOK {
		t.Fatalf("expected 200 OK for HTML statement, got %d", wStmtHTML.Code)
	}
	if !strings.Contains(wStmtHTML.Body.String(), "AI Usage &amp; Chargeback Statement") {
		t.Errorf("expected statement HTML output")
	}

	// 5. POST /api/simulator/simulate
	simBody := `{"subscription_increase_usd":10.0,"model_replacements":{"gpt-4o":"gpt-4o-mini"}}`
	reqSim := httptest.NewRequest(http.MethodPost, "/api/simulator/simulate", strings.NewReader(simBody))
	wSim := httptest.NewRecorder()
	mux.ServeHTTP(wSim, reqSim)

	if wSim.Code != http.StatusOK {
		t.Fatalf("expected 200 OK for simulator, got %d", wSim.Code)
	}
	var simResp SimulationResponse
	_ = json.Unmarshal(wSim.Body.Bytes(), &simResp)
	if simResp.SimulatedRevenueUSD != 109.0 {
		t.Errorf("expected simulated revenue 109.0, got %f", simResp.SimulatedRevenueUSD)
	}

	// 6. GET /metrics (Prometheus)
	reqMetrics := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	wMetrics := httptest.NewRecorder()
	mux.ServeHTTP(wMetrics, reqMetrics)

	if wMetrics.Code != http.StatusOK {
		t.Fatalf("expected 200 OK for /metrics, got %d", wMetrics.Code)
	}
	if !strings.Contains(wMetrics.Body.String(), "llmmeter_http_requests_total") {
		t.Errorf("expected prometheus metrics format")
	}
}
