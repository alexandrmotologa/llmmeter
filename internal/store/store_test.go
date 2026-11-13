package store

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestStoreEndToEnd(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "llmmeter-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	dbPath := filepath.Join(tmpDir, "finops.db")
	s, err := New(dbPath)
	if err != nil {
		t.Fatalf("failed to init store: %v", err)
	}
	defer s.Close()

	// 1. Create Tenants
	budget := 50.0
	err = s.CreateTenant(Tenant{
		ID:               "tenant-alpha",
		Name:             "Alpha Corp",
		MonthlyPlanUSD:   29.0, // Pays $29/month
		MonthlyBudgetUSD: &budget,
	})
	if err != nil {
		t.Fatalf("create tenant alpha failed: %v", err)
	}

	err = s.CreateTenant(Tenant{
		ID:             "tenant-beta",
		Name:           "Beta Enterprise",
		MonthlyPlanUSD: 499.0, // Pays $499/month
	})
	if err != nil {
		t.Fatalf("create tenant beta failed: %v", err)
	}

	// 2. Log Requests
	now := time.Now()
	// Alpha runs expensive queries -> $83 in tokens! (Silent margin bleed)
	for i := 0; i < 10; i++ {
		err = s.LogRequest(RequestLog{
			TenantID:         "tenant-alpha",
			FeatureTag:       "smart-summary",
			Provider:         "anthropic",
			Model:            "claude-3-5-sonnet",
			PromptTokens:     10000,
			CompletionTokens: 5000,
			CostUSD:          8.30, // 10 * 8.30 = $83.00
			LatencyMS:        450,
			Timestamp:        now.Add(-time.Duration(i) * time.Hour),
		})
		if err != nil {
			t.Fatalf("log request failed: %v", err)
		}
	}

	// Beta runs moderate queries -> $40 total cost
	err = s.LogRequest(RequestLog{
		TenantID:         "tenant-beta",
		FeatureTag:       "chat-copilot",
		Provider:         "openai",
		Model:            "gpt-4o-mini",
		PromptTokens:     50000,
		CompletionTokens: 20000,
		CostUSD:          40.0,
		LatencyMS:        120,
		Timestamp:        now,
	})
	if err != nil {
		t.Fatalf("log request failed: %v", err)
	}

	// 3. Test Unit Economics & Margin Calculation
	metrics, err := s.GetTenantMetrics("")
	if err != nil {
		t.Fatalf("GetTenantMetrics failed: %v", err)
	}
	if len(metrics) != 2 {
		t.Fatalf("expected 2 tenant metrics, got %d", len(metrics))
	}

	// Find Alpha
	var alpha, beta *TenantMetric
	for i := range metrics {
		if metrics[i].TenantID == "tenant-alpha" {
			alpha = &metrics[i]
		}
		if metrics[i].TenantID == "tenant-beta" {
			beta = &metrics[i]
		}
	}

	if alpha == nil || beta == nil {
		t.Fatalf("expected both alpha and beta metrics")
	}

	// Alpha paid $29, spent $83 -> Gross Margin = 29 - 83 = -$54 (negative margin!)
	if !alpha.IsUnprofitable {
		t.Errorf("expected Alpha to be unprofitable, got IsUnprofitable = false")
	}
	if alpha.GrossMarginUSD >= 0 {
		t.Errorf("expected negative gross margin USD for Alpha, got %f", alpha.GrossMarginUSD)
	}
	if alpha.TotalCostUSD < 80 {
		t.Errorf("expected TotalCostUSD ~83 for Alpha, got %f", alpha.TotalCostUSD)
	}

	// Beta paid $499, spent $40 -> Gross Margin = $459 (healthy profitable customer!)
	if beta.IsUnprofitable {
		t.Errorf("expected Beta to be profitable, got IsUnprofitable = true")
	}
	if beta.GrossMarginUSD <= 0 {
		t.Errorf("expected positive gross margin for Beta, got %f", beta.GrossMarginUSD)
	}

	// 4. Test Unprofitable Tenants Filter
	unprofitable, err := s.GetUnprofitableTenants("")
	if err != nil {
		t.Fatalf("GetUnprofitableTenants failed: %v", err)
	}
	if len(unprofitable) != 1 || unprofitable[0].TenantID != "tenant-alpha" {
		t.Errorf("expected exactly tenant-alpha in unprofitable list, got %v", unprofitable)
	}

	// 5. Test Feature Breakdown
	features, err := s.GetFeatureBreakdown("tenant-alpha", "")
	if err != nil {
		t.Fatalf("GetFeatureBreakdown failed: %v", err)
	}
	if len(features) == 0 || features[0].FeatureTag != "smart-summary" {
		t.Errorf("expected smart-summary feature breakdown, got %v", features)
	}

	// 6. Test Alerts & Recommendations
	err = s.CreateAlert(Alert{
		TenantID: "tenant-alpha",
		Type:     "negative_margin",
		Message:  "Tenant Alpha consumed $83 vs $29 revenue",
		Severity: "critical",
	})
	if err != nil {
		t.Fatalf("CreateAlert failed: %v", err)
	}

	alerts, err := s.ListAlerts(nil)
	if err != nil || len(alerts) != 1 {
		t.Fatalf("ListAlerts failed: %v, count=%d", err, len(alerts))
	}

	err = s.ResolveAlert(alerts[0].ID)
	if err != nil {
		t.Fatalf("ResolveAlert failed: %v", err)
	}

	resolvedFalse := false
	unresolved, err := s.ListAlerts(&resolvedFalse)
	if err != nil || len(unresolved) != 0 {
		t.Fatalf("expected 0 unresolved alerts, got %d", len(unresolved))
	}

	// 7. Test Recommendations
	err = s.CreateRecommendation(Recommendation{
		TenantID:          "tenant-alpha",
		FeatureTag:        "smart-summary",
		CurrentModel:      "claude-3-5-sonnet",
		SuggestedModel:    "claude-3-5-haiku",
		MonthlySavingsUSD: 58.10,
		Reason:            "Route with negative gross margin. Claude 3.5 Haiku reduces costs by ~70% with comparable summary quality.",
	})
	if err != nil {
		t.Fatalf("CreateRecommendation failed: %v", err)
	}

	recs, err := s.ListRecommendations(nil)
	if err != nil || len(recs) != 1 {
		t.Fatalf("ListRecommendations failed: %v", err)
	}
	if recs[0].MonthlySavingsUSD < 50 {
		t.Errorf("expected monthly savings > 50, got %f", recs[0].MonthlySavingsUSD)
	}
}
