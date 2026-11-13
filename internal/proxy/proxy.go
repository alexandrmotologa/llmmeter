package proxy

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/alexandrmotologa/llmmeter/internal/finops"
	"github.com/alexandrmotologa/llmmeter/internal/pricing"
	"github.com/alexandrmotologa/llmmeter/internal/store"
	"github.com/google/uuid"
)

var keyPrefixRegex = regexp.MustCompile(`^sk-t([^-]+(?:-[^-]+)*?)-?(sk-.+)$`)

// Context holds tenant and attribution metadata extracted from request headers.
type Context struct {
	TenantID   string
	FeatureTag string
	UserID     string
	RealAPIKey string
}

// Handler is the reverse proxy HTTP handler.
type Handler struct {
	store *store.Store
}

// NewHandler creates a new proxy handler.
func NewHandler(s *store.Store) *Handler {
	return &Handler{store: s}
}

// ExtractContext extracts tenant attribution details from incoming request.
func ExtractContext(r *http.Request) Context {
	tenantID := strings.TrimSpace(r.Header.Get("X-Tenant-ID"))
	featureTag := strings.TrimSpace(r.Header.Get("X-Feature-Tag"))
	if featureTag == "" {
		featureTag = strings.TrimSpace(r.Header.Get("X-Feature-ID"))
	}
	if featureTag == "" {
		featureTag = "default"
	}
	userID := strings.TrimSpace(r.Header.Get("X-User-ID"))

	authHeader := r.Header.Get("Authorization")
	rawKey := strings.TrimSpace(strings.TrimPrefix(authHeader, "Bearer "))

	// Check if API key carries tenant prefix: sk-t{tenantId}-{realKey}
	if tenantID == "" && rawKey != "" {
		matches := keyPrefixRegex.FindStringSubmatch(rawKey)
		if len(matches) == 3 {
			tenantID = matches[1]
			rawKey = matches[2]
		}
	}

	if tenantID == "" {
		tenantID = "anonymous"
	}

	return Context{
		TenantID:   tenantID,
		FeatureTag: featureTag,
		UserID:     userID,
		RealAPIKey: rawKey,
	}
}

type chatRequest struct {
	Model         string          `json:"model"`
	Stream        bool            `json:"stream"`
	StreamOptions *map[string]any `json:"stream_options,omitempty"`
}

type tokenUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
	PromptDetails    *struct {
		CachedTokens int `json:"cached_tokens"`
	} `json:"prompt_tokens_details,omitempty"`

	// Anthropic usage field names
	InputTokens          int `json:"input_tokens"`
	OutputTokens         int `json:"output_tokens"`
	CacheReadInputTokens int `json:"cache_read_input_tokens"`
}

func (u tokenUsage) Prompt() int {
	if u.PromptTokens > 0 {
		return u.PromptTokens
	}
	return u.InputTokens
}

func (u tokenUsage) Completion() int {
	if u.CompletionTokens > 0 {
		return u.CompletionTokens
	}
	return u.OutputTokens
}

func (u tokenUsage) Cached() int {
	if u.PromptDetails != nil && u.PromptDetails.CachedTokens > 0 {
		return u.PromptDetails.CachedTokens
	}
	return u.CacheReadInputTokens
}

func buildUpstreamAuth(provider, clientKey string) string {
	if clientKey != "" && !strings.HasPrefix(clientKey, "dummy") {
		return "Bearer " + clientKey
	}
	switch provider {
	case "openai":
		return "Bearer " + os.Getenv("OPENAI_API_KEY")
	case "anthropic":
		return "Bearer " + os.Getenv("ANTHROPIC_API_KEY")
	case "gemini":
		return "Bearer " + os.Getenv("GEMINI_API_KEY")
	case "groq":
		return "Bearer " + os.Getenv("GROQ_API_KEY")
	case "mistral":
		return "Bearer " + os.Getenv("MISTRAL_API_KEY")
	default:
		return "Bearer " + os.Getenv("OPENAI_API_KEY")
	}
}

// ServeChatCompletions handles POST /v1/chat/completions with token metering and SSE pass-through.
func (h *Handler) ServeChatCompletions(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, `{"error":{"message":"Method not allowed"}}`, http.StatusMethodNotAllowed)
		return
	}

	start := time.Now()
	ctx := ExtractContext(r)

	// Ensure tenant exists in database
	tenant, _ := h.store.GetTenant(ctx.TenantID)
	if tenant == nil && ctx.TenantID != "" {
		_ = h.store.CreateTenant(store.Tenant{
			ID:             ctx.TenantID,
			Name:           ctx.TenantID,
			MonthlyPlanUSD: 0.0,
		})
		tenant, _ = h.store.GetTenant(ctx.TenantID)
	}

	// Read body
	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, `{"error":{"message":"Failed to read request body"}}`, http.StatusBadRequest)
		return
	}

	var req chatRequest
	if err := json.Unmarshal(bodyBytes, &req); err != nil {
		http.Error(w, `{"error":{"message":"Invalid JSON payload"}}`, http.StatusBadRequest)
		return
	}

	model := req.Model
	if model == "" {
		model = "gpt-4o-mini"
	}

	provider := r.Header.Get("X-Provider")
	if provider == "" {
		provider = pricing.DetectProvider(model)
	}

	// Budget Guardrail
	if tenant != nil && tenant.MonthlyBudgetUSD != nil && *tenant.MonthlyBudgetUSD > 0 {
		exceeded, _, _, _ := finops.CheckBudgetCap(h.store, tenant, 0.01)
		if exceeded {
			w.Header().Set("Content-Type", "application/json")
			w.Header().Set("X-Budget-Exceeded", "true")
			w.WriteHeader(http.StatusTooManyRequests)
			_, _ = w.Write([]byte(`{"error":{"message":"Monthly LLM budget cap exceeded for this tenant.","type":"budget_exceeded"}}`))
			return
		}
	}

	// If streaming, ensure stream_options.include_usage = true for OpenAI-compatible providers
	if req.Stream {
		var genericMap map[string]any
		if err := json.Unmarshal(bodyBytes, &genericMap); err == nil {
			if _, ok := genericMap["stream_options"]; !ok {
				genericMap["stream_options"] = map[string]any{"include_usage": true}
				if modified, err := json.Marshal(genericMap); err == nil {
					bodyBytes = modified
				}
			}
		}
	}

	baseURL := pricing.ProviderBaseURL(provider)
	targetURL, err := url.Parse(baseURL + "/chat/completions")
	if err != nil {
		http.Error(w, `{"error":{"message":"Failed to resolve upstream URL"}}`, http.StatusInternalServerError)
		return
	}

	// Prepare upstream request
	upstreamReq, err := http.NewRequestWithContext(r.Context(), http.MethodPost, targetURL.String(), bytes.NewReader(bodyBytes))
	if err != nil {
		http.Error(w, `{"error":{"message":"Failed to create upstream request"}}`, http.StatusInternalServerError)
		return
	}

	// Copy forwardable headers
	for k, vv := range r.Header {
		lower := strings.ToLower(k)
		if lower == "content-type" || lower == "accept" || lower == "x-request-id" || lower == "anthropic-version" {
			for _, v := range vv {
				upstreamReq.Header.Add(k, v)
			}
		}
	}
	upstreamReq.Header.Set("Authorization", buildUpstreamAuth(provider, ctx.RealAPIKey))
	if provider == "anthropic" && upstreamReq.Header.Get("anthropic-version") == "" {
		upstreamReq.Header.Set("anthropic-version", "2023-06-01")
	}

	// Execute upstream
	client := &http.Client{Timeout: 5 * time.Minute}
	upstreamResp, err := client.Do(upstreamReq)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error":{"message":"Upstream provider unreachable: %s"}}`, err.Error()), http.StatusBadGateway)
		return
	}
	defer upstreamResp.Body.Close()

	// Forward response headers
	for k, vv := range upstreamResp.Header {
		for _, v := range vv {
			w.Header().Add(k, v)
		}
	}
	w.WriteHeader(upstreamResp.StatusCode)

	var promptTokens, completionTokens, cachedTokens int

	if req.Stream {
		// Flush SSE events in real-time
		flusher, ok := w.(http.Flusher)
		reader := bufio.NewReader(upstreamResp.Body)

		for {
			line, err := reader.ReadBytes('\n')
			if len(line) > 0 {
				_, _ = w.Write(line)
				if ok {
					flusher.Flush()
				}

				// Inspect line for SSE usage block
				lineStr := strings.TrimSpace(string(line))
				if strings.HasPrefix(lineStr, "data: ") {
					dataStr := strings.TrimPrefix(lineStr, "data: ")
					if dataStr != "[DONE]" {
						var parsed struct {
							Usage *tokenUsage `json:"usage"`
						}
						if err := json.Unmarshal([]byte(dataStr), &parsed); err == nil && parsed.Usage != nil {
							promptTokens = parsed.Usage.Prompt()
							completionTokens = parsed.Usage.Completion()
							cachedTokens = parsed.Usage.Cached()
						}
					}
				}
			}
			if err != nil {
				break
			}
		}
	} else {
		// Non-streaming: copy and parse usage
		respBytes, err := io.ReadAll(upstreamResp.Body)
		if err == nil {
			_, _ = w.Write(respBytes)

			// Decompress if gzip
			var decompressed []byte = respBytes
			if upstreamResp.Header.Get("Content-Encoding") == "gzip" {
				if gr, err := gzip.NewReader(bytes.NewReader(respBytes)); err == nil {
					decompressed, _ = io.ReadAll(gr)
					_ = gr.Close()
				}
			}

			var parsed struct {
				Usage *tokenUsage `json:"usage"`
			}
			if err := json.Unmarshal(decompressed, &parsed); err == nil && parsed.Usage != nil {
				promptTokens = parsed.Usage.Prompt()
				completionTokens = parsed.Usage.Completion()
				cachedTokens = parsed.Usage.Cached()
			}
		}
	}

	latency := time.Since(start).Milliseconds()

	// If provider didn't return usage (e.g. streaming without include_usage or mock), estimate from lengths
	if promptTokens == 0 && completionTokens == 0 {
		promptTokens = len(bodyBytes) / 4
		completionTokens = 50
	}

	costUSD := pricing.ComputeCost(model, promptTokens, completionTokens, cachedTokens)

	// Asynchronous attribution & margin checking (never slows down client response!)
	go func() {
		logEntry := store.RequestLog{
			ID:               uuid.New().String(),
			TenantID:         ctx.TenantID,
			FeatureTag:       ctx.FeatureTag,
			UserID:           ctx.UserID,
			Provider:         provider,
			Model:            model,
			PromptTokens:     promptTokens,
			CompletionTokens: completionTokens,
			CachedTokens:     cachedTokens,
			CostUSD:          costUSD,
			LatencyMS:        latency,
			Timestamp:        time.Now(),
		}
		_ = h.store.LogRequest(logEntry)

		// Anomaly & margin audit
		_, _ = finops.CheckRequestAnomaly(h.store, ctx.TenantID, ctx.FeatureTag, model, costUSD)
		if tenant != nil {
			_, _ = finops.CheckNegativeMargin(h.store, tenant)
		}
	}()
}

// ServeEmbeddings handles POST /v1/embeddings.
func (h *Handler) ServeEmbeddings(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, `{"error":{"message":"Method not allowed"}}`, http.StatusMethodNotAllowed)
		return
	}

	start := time.Now()
	ctx := ExtractContext(r)

	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, `{"error":{"message":"Failed to read body"}}`, http.StatusBadRequest)
		return
	}

	var req struct {
		Model string `json:"model"`
	}
	_ = json.Unmarshal(bodyBytes, &req)
	model := req.Model
	if model == "" {
		model = "text-embedding-3-small"
	}

	provider := pricing.DetectProvider(model)
	baseURL := pricing.ProviderBaseURL(provider)

	targetURL, _ := url.Parse(baseURL + "/embeddings")
	upstreamReq, err := http.NewRequestWithContext(r.Context(), http.MethodPost, targetURL.String(), bytes.NewReader(bodyBytes))
	if err != nil {
		http.Error(w, `{"error":{"message":"Upstream init error"}}`, http.StatusInternalServerError)
		return
	}

	upstreamReq.Header.Set("Content-Type", "application/json")
	upstreamReq.Header.Set("Authorization", buildUpstreamAuth(provider, ctx.RealAPIKey))

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(upstreamReq)
	if err != nil {
		http.Error(w, `{"error":{"message":"Upstream unreachable"}}`, http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	respBytes, _ := io.ReadAll(resp.Body)
	for k, vv := range resp.Header {
		for _, v := range vv {
			w.Header().Add(k, v)
		}
	}
	w.WriteHeader(resp.StatusCode)
	_, _ = w.Write(respBytes)

	var parsed struct {
		Usage struct {
			PromptTokens int `json:"prompt_tokens"`
			TotalTokens  int `json:"total_tokens"`
		} `json:"usage"`
	}
	_ = json.Unmarshal(respBytes, &parsed)
	promptTokens := parsed.Usage.PromptTokens
	if promptTokens == 0 {
		promptTokens = len(bodyBytes) / 4
	}

	costUSD := pricing.ComputeCost(model, promptTokens, 0, 0)
	latency := time.Since(start).Milliseconds()

	go func() {
		_ = h.store.LogRequest(store.RequestLog{
			TenantID:         ctx.TenantID,
			FeatureTag:       ctx.FeatureTag,
			UserID:           ctx.UserID,
			Provider:         provider,
			Model:            model,
			PromptTokens:     promptTokens,
			CompletionTokens: 0,
			CostUSD:          costUSD,
			LatencyMS:        latency,
			Timestamp:        time.Now(),
		})
	}()
}

// ServeModels returns model list with cost metadata.
func (h *Handler) ServeModels(w http.ResponseWriter, r *http.Request) {
	all := pricing.ListAllPricing()
	type ModelData struct {
		ID       string `json:"id"`
		Object   string `json:"object"`
		Provider string `json:"provider"`
		Pricing  struct {
			InputPer1M       float64 `json:"input_per_1m_tokens"`
			OutputPer1M      float64 `json:"output_per_1m_tokens"`
			CachedInputPer1M float64 `json:"cached_input_per_1m_tokens"`
		} `json:"pricing"`
	}

	data := make([]ModelData, len(all))
	for i, p := range all {
		data[i] = ModelData{
			ID:       p.Model,
			Object:   "model",
			Provider: p.Provider,
		}
		data[i].Pricing.InputPer1M = p.InputPer1M
		data[i].Pricing.OutputPer1M = p.OutputPer1M
		data[i].Pricing.CachedInputPer1M = p.CachedInputPer1M
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"object": "list",
		"data":   data,
	})
}

// ServeHealth returns health status.
func (h *Handler) ServeHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"status":    "ok",
		"engine":    "llmmeter-go",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	})
}

// PassthroughReverseProxy creates a generic reverse proxy if needed.
func PassthroughReverseProxy(target *url.URL) *httputil.ReverseProxy {
	proxy := httputil.NewSingleHostReverseProxy(target)
	proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		http.Error(w, fmt.Sprintf(`{"error":{"message":"Proxy error: %s"}}`, err.Error()), http.StatusBadGateway)
	}
	return proxy
}
