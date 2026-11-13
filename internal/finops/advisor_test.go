package finops

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/alexandrmotologa/llmmeter/internal/store"
)

func TestFinopsAdvisor(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "llmmeter-finops-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	dbPath := filepath.Join(tmpDir, "finops.db")
	s, err := store.New(dbPath)
	if err != nil {
		t.Fatalf("failed to init store: %v", err)
	}
	defer s.Close()

	// Setup tenant with $29 plan and $50 budget
	budget := 50.0
	tenant := store.Tenant{
		ID:               "corp-gamma",
		Name:             "Gamma Analytics",
		MonthlyPlanUSD:   29.0,
		MonthlyBudgetUSD: &budget,
	}
	_ = s.CreateTenant(tenant)

	// Baseline queries for anomaly check
	now := time.Now()
	for i := 0; i < 6; i++ {
		_ = s.LogRequest(store.RequestLog{
			TenantID:         "corp-gamma",
			FeatureTag:       "search",
			Provider:         "openai",
			Model:            "gpt-4o",
			PromptTokens:     1000,
			CompletionTokens: 200,
			CostUSD:          0.0045, // small baseline
			LatencyMS:        200,
			Timestamp:        now.Add(-time.Duration(i+1) * time.Hour),
		})
	}

	// 1. Test Anomaly Detection: a request of $0.50 (100x baseline!)
	alert, err := CheckRequestAnomaly(s, "corp-gamma", "search", "gpt-4o", 0.50)
	if err != nil {
		t.Fatalf("CheckRequestAnomaly error: %v", err)
	}
	if alert == nil {
		t.Errorf("expected anomaly alert, got nil")
	} else if alert.Type != "anomaly" {
		t.Errorf("expected alert type 'anomaly', got %s", alert.Type)
	}

	// 2. Test Budget Cap
	// Add $60 spend to exceed the $50 budget
	_ = s.LogRequest(store.RequestLog{
		TenantID:         "corp-gamma",
		FeatureTag:       "heavy-batch",
		Provider:         "anthropic",
		Model:            "claude-3-5-sonnet",
		PromptTokens:     500000,
		CompletionTokens: 100000,
		CostUSD:          60.00,
		LatencyMS:        800,
		Timestamp:        now,
	})

	exceeded, spend, bAlert, err := CheckBudgetCap(s, &tenant, 1.0)
	if err != nil {
		t.Fatalf("CheckBudgetCap error: %v", err)
	}
	if !exceeded {
		t.Errorf("expected budget cap exceeded = true, got false (spend: %f)", spend)
	}
	if bAlert == nil || bAlert.Type != "budget" {
		t.Errorf("expected budget alert")
	}

	// 3. Test Negative Margin Alert (pays $29, spent > $60)
	mAlert, err := CheckNegativeMargin(s, &tenant)
	if err != nil {
		t.Fatalf("CheckNegativeMargin error: %v", err)
	}
	if mAlert == nil || mAlert.Type != "negative_margin" {
		t.Errorf("expected negative margin alert, got %v", mAlert)
	}

	// 4. Test Automated Recommendations
	recs, err := GenerateRecommendations(s, "")
	if err != nil {
		t.Fatalf("GenerateRecommendations error: %v", err)
	}
	if len(recs) == 0 {
		t.Errorf("expected at least 1 recommendation for heavy claude-3-5-sonnet route, got 0")
	} else {
		first := recs[0]
		if !strings.Contains(first.CurrentModel, "sonnet") || !strings.Contains(first.SuggestedModel, "haiku") {
			t.Errorf("unexpected models in recommendation: %s -> %s", first.CurrentModel, first.SuggestedModel)
		}
		if first.MonthlySavingsUSD <= 0 {
			t.Errorf("expected positive monthly savings, got %f", first.MonthlySavingsUSD)
		}
	}
}
