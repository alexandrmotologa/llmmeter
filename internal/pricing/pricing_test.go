package pricing

import (
	"math"
	"testing"
)

func TestDetectProvider(t *testing.T) {
	tests := []struct {
		model    string
		expected string
	}{
		{"gpt-4o", "openai"},
		{"o1-preview", "openai"},
		{"claude-3-5-sonnet-20241022", "anthropic"},
		{"gemini-1.5-flash", "gemini"},
		{"llama-3.1-70b-versatile", "groq"},
		{"mistral-large-latest", "mistral"},
		{"bedrock/anthropic.claude-3-5-sonnet", "bedrock"},
	}

	for _, tt := range tests {
		got := DetectProvider(tt.model)
		if got != tt.expected {
			t.Errorf("DetectProvider(%q) = %q; want %q", tt.model, got, tt.expected)
		}
	}
}

func TestComputeCost(t *testing.T) {
	// gpt-4o: 2.50 input per 1M, 10.00 output per 1M, 1.25 cached input per 1M
	// 1000 prompt tokens (with 200 cached) and 500 completion tokens
	// regularPrompt = 800 tokens = 800/1,000,000 * 2.50 = 0.002
	// cachedPrompt = 200 tokens = 200/1,000,000 * 1.25 = 0.00025
	// completion = 500 tokens = 500/1,000,000 * 10.00 = 0.005
	// total = 0.00725
	cost := ComputeCost("gpt-4o", 1000, 500, 200)
	expected := 0.00725
	if math.Abs(cost-expected) > 0.000001 {
		t.Errorf("ComputeCost() = %f; want %f", cost, expected)
	}
}

func TestGetPricingFallback(t *testing.T) {
	p := GetPricing("claude-3-5-sonnet-future-version")
	if p.Provider != "anthropic" {
		t.Errorf("Expected provider anthropic, got %s", p.Provider)
	}
	if p.InputPer1M <= 0 {
		t.Errorf("Expected positive InputPer1M, got %f", p.InputPer1M)
	}
}
