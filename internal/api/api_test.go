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

	"github.com/alexandrmotologa/llmmeter/internal/store"
)

func TestAPIServer(t *testing.T) {
	tmpDir, _ := os.MkdirTemp("", "llmmeter-api-test-*")
	defer os.RemoveAll(tmpDir)

	s, _ := store.New(filepath.Join(tmpDir, "finops.db"))
	defer s.Close()

	// Seed one tenant and log
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
	var ov store.OverviewSummary
	_ = json.Unmarshal(wOverview.Body.Bytes(), &ov)
	if ov.ActiveTenantsCount != 1 || ov.TotalRevenueUSD != 99.0 {
		t.Errorf("unexpected overview data: %+v", ov)
	}

	// 2. GET /api/tenants
	reqTenants := httptest.NewRequest(http.MethodGet, "/api/tenants", nil)
	wTenants := httptest.NewRecorder()
	mux.ServeHTTP(wTenants, reqTenants)

	if wTenants.Code != http.StatusOK {
		t.Fatalf("expected 200 OK for tenants, got %d", wTenants.Code)
	}

	// 3. POST /api/tenants
	newTenantJSON := `{"id":"tenant-epsilon","name":"Epsilon Inc","monthly_plan_usd":199.0}`
	reqPostTenant := httptest.NewRequest(http.MethodPost, "/api/tenants", strings.NewReader(newTenantJSON))
	wPostTenant := httptest.NewRecorder()
	mux.ServeHTTP(wPostTenant, reqPostTenant)

	if wPostTenant.Code != http.StatusCreated {
		t.Fatalf("expected 201 Created for post tenant, got %d", wPostTenant.Code)
	}

	// 4. GET /api/features
	reqFeatures := httptest.NewRequest(http.MethodGet, "/api/features", nil)
	wFeatures := httptest.NewRecorder()
	mux.ServeHTTP(wFeatures, reqFeatures)

	if wFeatures.Code != http.StatusOK {
		t.Fatalf("expected 200 OK for features, got %d", wFeatures.Code)
	}

	// 5. GET /api/export/csv
	reqCSV := httptest.NewRequest(http.MethodGet, "/api/export/csv", nil)
	wCSV := httptest.NewRecorder()
	mux.ServeHTTP(wCSV, reqCSV)

	if wCSV.Code != http.StatusOK {
		t.Fatalf("expected 200 OK for csv export, got %d", wCSV.Code)
	}
	if !strings.Contains(wCSV.Body.String(), "Delta Corp") {
		t.Errorf("expected CSV to contain Delta Corp, got: %s", wCSV.Body.String())
	}
}
