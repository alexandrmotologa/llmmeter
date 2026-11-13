package proxy

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/alexandrmotologa/llmmeter/internal/store"
)

func TestExtractContext(t *testing.T) {
	// 1. Explicit headers
	req := httptest.NewRequest(http.MethodPost, "/v1/chat/completions", nil)
	req.Header.Set("X-Tenant-ID", "tenant-123")
	req.Header.Set("X-Feature-Tag", "copilot-v2")
	req.Header.Set("X-User-ID", "user-456")
	req.Header.Set("Authorization", "Bearer sk-test-key")

	ctx := ExtractContext(req)
	if ctx.TenantID != "tenant-123" {
		t.Errorf("expected tenant-123, got %s", ctx.TenantID)
	}
	if ctx.FeatureTag != "copilot-v2" {
		t.Errorf("expected copilot-v2, got %s", ctx.FeatureTag)
	}
	if ctx.UserID != "user-456" {
		t.Errorf("expected user-456, got %s", ctx.UserID)
	}
	if ctx.RealAPIKey != "sk-test-key" {
		t.Errorf("expected sk-test-key, got %s", ctx.RealAPIKey)
	}

	// 2. Prefixed API key: sk-tacme-corp-sk-livekey123
	req2 := httptest.NewRequest(http.MethodPost, "/v1/chat/completions", nil)
	req2.Header.Set("Authorization", "Bearer sk-tacme-corp-sk-livekey123")
	ctx2 := ExtractContext(req2)
	if ctx2.TenantID != "acme-corp" {
		t.Errorf("expected acme-corp from key prefix, got %s", ctx2.TenantID)
	}
	if ctx2.RealAPIKey != "sk-livekey123" {
		t.Errorf("expected extracted real key sk-livekey123, got %s", ctx2.RealAPIKey)
	}
}

func TestProxyHealthAndModels(t *testing.T) {
	tmpDir, _ := os.MkdirTemp("", "llmmeter-proxy-test-*")
	defer os.RemoveAll(tmpDir)

	s, _ := store.New(filepath.Join(tmpDir, "finops.db"))
	defer s.Close()

	h := NewHandler(s)

	// Health
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()
	h.ServeHealth(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200 OK for health, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "llmmeter-go") {
		t.Errorf("unexpected health body: %s", w.Body.String())
	}

	// Models
	reqModels := httptest.NewRequest(http.MethodGet, "/v1/models", nil)
	wModels := httptest.NewRecorder()
	h.ServeModels(wModels, reqModels)

	if wModels.Code != http.StatusOK {
		t.Errorf("expected 200 OK for models, got %d", wModels.Code)
	}
	if !strings.Contains(wModels.Body.String(), "gpt-4o") {
		t.Errorf("expected gpt-4o in models list")
	}
}

func TestProxyBudgetGuardrail(t *testing.T) {
	tmpDir, _ := os.MkdirTemp("", "llmmeter-proxy-budget-*")
	defer os.RemoveAll(tmpDir)

	s, _ := store.New(filepath.Join(tmpDir, "finops.db"))
	defer s.Close()

	// Tenant with $10 budget that already spent $15
	budget := 10.0
	_ = s.CreateTenant(store.Tenant{
		ID:               "tenant-broke",
		Name:             "Broke Startup",
		MonthlyPlanUSD:   29.0,
		MonthlyBudgetUSD: &budget,
	})
	_ = s.LogRequest(store.RequestLog{
		TenantID:         "tenant-broke",
		FeatureTag:       "heavy-ocr",
		Provider:         "openai",
		Model:            "gpt-4o",
		PromptTokens:     100000,
		CompletionTokens: 50000,
		CostUSD:          15.00,
		LatencyMS:        300,
		Timestamp:        time.Now(),
	})

	h := NewHandler(s)

	body := `{"model":"gpt-4o-mini","messages":[{"role":"user","content":"hello"}]}`
	req := httptest.NewRequest(http.MethodPost, "/v1/chat/completions", strings.NewReader(body))
	req.Header.Set("X-Tenant-ID", "tenant-broke")
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.ServeChatCompletions(w, req)

	if w.Code != http.StatusTooManyRequests {
		t.Errorf("expected 429 Too Many Requests for budget cap breach, got %d (body: %s)", w.Code, w.Body.String())
	}
	if w.Header().Get("X-Budget-Exceeded") != "true" {
		t.Errorf("expected X-Budget-Exceeded: true header")
	}
}

func TestTokenUsageMethods(t *testing.T) {
	u := tokenUsage{
		PromptTokens:     150,
		CompletionTokens: 50,
		TotalTokens:      200,
	}
	if u.Prompt() != 150 || u.Completion() != 50 {
		t.Errorf("unexpected token counts: prompt=%d, completion=%d", u.Prompt(), u.Completion())
	}

	// Anthropic usage
	u2 := tokenUsage{
		InputTokens:          250,
		OutputTokens:         100,
		CacheReadInputTokens: 50,
	}
	if u2.Prompt() != 250 || u2.Completion() != 100 || u2.Cached() != 50 {
		t.Errorf("unexpected anthropic token counts")
	}
}
