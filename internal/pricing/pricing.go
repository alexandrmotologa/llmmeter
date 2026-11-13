package pricing

import (
	"strings"
	"sync"
)

// ModelPricing represents pricing in USD per 1,000,000 tokens.
type ModelPricing struct {
	Provider          string  `json:"provider"`
	Model             string  `json:"model"`
	InputPer1M        float64 `json:"input_per_1m"`
	OutputPer1M       float64 `json:"output_per_1m"`
	CachedInputPer1M  float64 `json:"cached_input_per_1m"`
}

var (
	mu      sync.RWMutex
	catalog = map[string]ModelPricing{
		// OpenAI
		"gpt-4o":                  {Provider: "openai", Model: "gpt-4o", InputPer1M: 2.50, OutputPer1M: 10.00, CachedInputPer1M: 1.25},
		"gpt-4o-2024-11-20":       {Provider: "openai", Model: "gpt-4o-2024-11-20", InputPer1M: 2.50, OutputPer1M: 10.00, CachedInputPer1M: 1.25},
		"gpt-4o-2024-08-06":       {Provider: "openai", Model: "gpt-4o-2024-08-06", InputPer1M: 2.50, OutputPer1M: 10.00, CachedInputPer1M: 1.25},
		"gpt-4o-mini":             {Provider: "openai", Model: "gpt-4o-mini", InputPer1M: 0.15, OutputPer1M: 0.60, CachedInputPer1M: 0.075},
		"gpt-4o-mini-2024-07-18":  {Provider: "openai", Model: "gpt-4o-mini-2024-07-18", InputPer1M: 0.15, OutputPer1M: 0.60, CachedInputPer1M: 0.075},
		"gpt-4-turbo":             {Provider: "openai", Model: "gpt-4-turbo", InputPer1M: 10.00, OutputPer1M: 30.00, CachedInputPer1M: 5.00},
		"gpt-4":                   {Provider: "openai", Model: "gpt-4", InputPer1M: 30.00, OutputPer1M: 60.00, CachedInputPer1M: 15.00},
		"gpt-3.5-turbo":           {Provider: "openai", Model: "gpt-3.5-turbo", InputPer1M: 0.50, OutputPer1M: 1.50, CachedInputPer1M: 0.25},
		"o1":                      {Provider: "openai", Model: "o1", InputPer1M: 15.00, OutputPer1M: 60.00, CachedInputPer1M: 7.50},
		"o1-mini":                 {Provider: "openai", Model: "o1-mini", InputPer1M: 3.00, OutputPer1M: 12.00, CachedInputPer1M: 1.50},
		"o3-mini":                 {Provider: "openai", Model: "o3-mini", InputPer1M: 1.10, OutputPer1M: 4.40, CachedInputPer1M: 0.55},
		"text-embedding-3-large":  {Provider: "openai", Model: "text-embedding-3-large", InputPer1M: 0.13, OutputPer1M: 0.0, CachedInputPer1M: 0.0},
		"text-embedding-3-small":  {Provider: "openai", Model: "text-embedding-3-small", InputPer1M: 0.02, OutputPer1M: 0.0, CachedInputPer1M: 0.0},

		// Anthropic
		"claude-3-5-sonnet-20241022": {Provider: "anthropic", Model: "claude-3-5-sonnet-20241022", InputPer1M: 3.00, OutputPer1M: 15.00, CachedInputPer1M: 0.30},
		"claude-3-5-sonnet":          {Provider: "anthropic", Model: "claude-3-5-sonnet", InputPer1M: 3.00, OutputPer1M: 15.00, CachedInputPer1M: 0.30},
		"claude-3-5-haiku-20241022":  {Provider: "anthropic", Model: "claude-3-5-haiku-20241022", InputPer1M: 1.00, OutputPer1M: 5.00, CachedInputPer1M: 0.10},
		"claude-3-5-haiku":           {Provider: "anthropic", Model: "claude-3-5-haiku", InputPer1M: 1.00, OutputPer1M: 5.00, CachedInputPer1M: 0.10},
		"claude-3-opus-20240229":     {Provider: "anthropic", Model: "claude-3-opus-20240229", InputPer1M: 15.00, OutputPer1M: 75.00, CachedInputPer1M: 1.50},
		"claude-3-opus":              {Provider: "anthropic", Model: "claude-3-opus", InputPer1M: 15.00, OutputPer1M: 75.00, CachedInputPer1M: 1.50},
		"claude-3-haiku-20240307":     {Provider: "anthropic", Model: "claude-3-haiku-20240307", InputPer1M: 0.25, OutputPer1M: 1.25, CachedInputPer1M: 0.03},

		// Google Gemini
		"gemini-1.5-pro":     {Provider: "gemini", Model: "gemini-1.5-pro", InputPer1M: 1.25, OutputPer1M: 5.00, CachedInputPer1M: 0.3125},
		"gemini-1.5-pro-002": {Provider: "gemini", Model: "gemini-1.5-pro-002", InputPer1M: 1.25, OutputPer1M: 5.00, CachedInputPer1M: 0.3125},
		"gemini-1.5-flash":   {Provider: "gemini", Model: "gemini-1.5-flash", InputPer1M: 0.075, OutputPer1M: 0.30, CachedInputPer1M: 0.01875},
		"gemini-1.5-flash-8b":{Provider: "gemini", Model: "gemini-1.5-flash-8b", InputPer1M: 0.0375, OutputPer1M: 0.15, CachedInputPer1M: 0.01},
		"gemini-2.0-flash":   {Provider: "gemini", Model: "gemini-2.0-flash", InputPer1M: 0.10, OutputPer1M: 0.40, CachedInputPer1M: 0.025},

		// Groq / Meta / Mistral
		"llama-3.1-70b-versatile": {Provider: "groq", Model: "llama-3.1-70b-versatile", InputPer1M: 0.59, OutputPer1M: 0.79, CachedInputPer1M: 0.30},
		"llama-3.1-70b":           {Provider: "groq", Model: "llama-3.1-70b", InputPer1M: 0.59, OutputPer1M: 0.79, CachedInputPer1M: 0.30},
		"llama-3.1-8b-instant":    {Provider: "groq", Model: "llama-3.1-8b-instant", InputPer1M: 0.05, OutputPer1M: 0.08, CachedInputPer1M: 0.025},
		"llama-3.1-8b":            {Provider: "groq", Model: "llama-3.1-8b", InputPer1M: 0.05, OutputPer1M: 0.08, CachedInputPer1M: 0.025},
		"mixtral-8x7b-32768":      {Provider: "groq", Model: "mixtral-8x7b-32768", InputPer1M: 0.24, OutputPer1M: 0.24, CachedInputPer1M: 0.12},
		"mistral-large-latest":    {Provider: "mistral", Model: "mistral-large-latest", InputPer1M: 2.00, OutputPer1M: 6.00, CachedInputPer1M: 1.00},
		"mistral-small-latest":    {Provider: "mistral", Model: "mistral-small-latest", InputPer1M: 0.20, OutputPer1M: 0.60, CachedInputPer1M: 0.10},
		"codestral-latest":        {Provider: "mistral", Model: "codestral-latest", InputPer1M: 0.20, OutputPer1M: 0.60, CachedInputPer1M: 0.10},

		// AWS Bedrock
		"bedrock/anthropic.claude-3-5-sonnet-20241022-v2:0": {Provider: "bedrock", Model: "bedrock/anthropic.claude-3-5-sonnet-20241022-v2:0", InputPer1M: 3.00, OutputPer1M: 15.00, CachedInputPer1M: 0.30},
		"bedrock/meta.llama3-1-70b-instruct-v1:0":           {Provider: "bedrock", Model: "bedrock/meta.llama3-1-70b-instruct-v1:0", InputPer1M: 0.99, OutputPer1M: 0.99, CachedInputPer1M: 0.50},
	}
)

// GetPricing retrieves pricing for the given model or matches the closest alias.
func GetPricing(model string) ModelPricing {
	mu.RLock()
	defer mu.RUnlock()

	normalized := strings.TrimSpace(strings.ToLower(model))
	if p, ok := catalog[normalized]; ok {
		return p
	}

	for k, v := range catalog {
		if strings.HasPrefix(normalized, k) || strings.HasPrefix(k, normalized) {
			return v
		}
	}

	// Fallback detection
	prov := DetectProvider(model)
	return ModelPricing{
		Provider:         prov,
		Model:            model,
		InputPer1M:       1.00,
		OutputPer1M:      3.00,
		CachedInputPer1M: 0.50,
	}
}

// ComputeCost calculates the cost in USD given prompt, completion, and cached tokens.
func ComputeCost(model string, promptTokens, completionTokens, cachedTokens int) float64 {
	p := GetPricing(model)

	regularPrompt := promptTokens - cachedTokens
	if regularPrompt < 0 {
		regularPrompt = 0
	}

	cost := (float64(regularPrompt) / 1_000_000.0 * p.InputPer1M) +
		(float64(cachedTokens) / 1_000_000.0 * p.CachedInputPer1M) +
		(float64(completionTokens) / 1_000_000.0 * p.OutputPer1M)

	return cost
}

// DetectProvider detects the provider from model naming patterns.
func DetectProvider(model string) string {
	m := strings.ToLower(model)
	switch {
	case strings.HasPrefix(m, "bedrock/"):
		return "bedrock"
	case strings.HasPrefix(m, "gpt-"), strings.HasPrefix(m, "o1"), strings.HasPrefix(m, "o3"), strings.HasPrefix(m, "text-embedding-"):
		return "openai"
	case strings.HasPrefix(m, "claude-"):
		return "anthropic"
	case strings.HasPrefix(m, "gemini-"):
		return "gemini"
	case strings.HasPrefix(m, "llama"), strings.HasPrefix(m, "mixtral"), strings.HasPrefix(m, "gemma"):
		return "groq"
	case strings.HasPrefix(m, "mistral"), strings.HasPrefix(m, "codestral"):
		return "mistral"
	default:
		return "openai"
	}
}

// ProviderBaseURL returns upstream base URL for a provider.
func ProviderBaseURL(provider string) string {
	switch provider {
	case "openai":
		return "https://api.openai.com/v1"
	case "anthropic":
		return "https://api.anthropic.com/v1"
	case "gemini":
		return "https://generativelanguage.googleapis.com/v1beta/openai"
	case "groq":
		return "https://api.groq.com/openai/v1"
	case "mistral":
		return "https://api.mistral.ai/v1"
	default:
		return "https://api.openai.com/v1"
	}
}

// ListAllPricing returns all registered model pricing records.
func ListAllPricing() []ModelPricing {
	mu.RLock()
	defer mu.RUnlock()

	result := make([]ModelPricing, 0, len(catalog))
	for _, p := range catalog {
		result = append(result, p)
	}
	return result
}

// RegisterOrUpdate allows updating or adding custom model pricing at runtime.
func RegisterOrUpdate(p ModelPricing) {
	mu.Lock()
	defer mu.Unlock()
	catalog[strings.TrimSpace(strings.ToLower(p.Model))] = p
}
