package api

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/alexandrmotologa/llmmeter/internal/finops"
	"github.com/alexandrmotologa/llmmeter/internal/metrics"
	"github.com/alexandrmotologa/llmmeter/internal/pricing"
	"github.com/alexandrmotologa/llmmeter/internal/store"
)

// Server provides the REST API handlers for the FinOps Studio.
type Server struct {
	store *store.Store
}

// NewServer creates a new API server.
func NewServer(s *store.Store) *Server {
	return &Server{store: s}
}

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]any{
		"error": map[string]string{"message": msg},
	})
}

// RegisterRoutes registers all API endpoints on the given multiplexer.
func (s *Server) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/api/overview", s.handleOverview)
	mux.HandleFunc("/api/tenants", s.handleTenants)
	mux.HandleFunc("/api/tenants/", s.handleTenantRoute)
	mux.HandleFunc("/api/features", s.handleFeatures)
	mux.HandleFunc("/api/models", s.handleModels)
	mux.HandleFunc("/api/alerts", s.handleAlerts)
	mux.HandleFunc("/api/alerts/", s.handleAlertAction)
	mux.HandleFunc("/api/recommendations", s.handleRecommendations)
	mux.HandleFunc("/api/recommendations/", s.handleRecommendationAction)
	mux.HandleFunc("/api/logs", s.handleLogs)
	mux.HandleFunc("/api/export/csv", s.handleExportCSV)
	mux.HandleFunc("/api/cache-analytics", s.handleCacheAnalytics)
	mux.HandleFunc("/api/simulator/simulate", s.handleSimulator)

	// Prometheus metrics exposition
	if m := metrics.Get(); m != nil {
		mux.HandleFunc("/metrics", m.Handler())
	}
}

func (s *Server) handleOverview(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}
	month := r.URL.Query().Get("month")
	summary, err := s.store.GetOverviewSummary(month)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, summary)
}

func (s *Server) handleTenants(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		month := r.URL.Query().Get("month")
		unprofitableOnly := r.URL.Query().Get("unprofitable") == "true"

		if unprofitableOnly {
			items, err := s.store.GetUnprofitableTenants(month)
			if err != nil {
				writeError(w, http.StatusInternalServerError, err.Error())
				return
			}
			writeJSON(w, http.StatusOK, items)
			return
		}

		metricsList, err := s.store.GetTenantMetrics(month)
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, metricsList)

	case http.MethodPost:
		var t store.Tenant
		if err := json.NewDecoder(r.Body).Decode(&t); err != nil {
			writeError(w, http.StatusBadRequest, "Invalid JSON payload")
			return
		}
		if t.Name == "" {
			writeError(w, http.StatusBadRequest, "Name is required")
			return
		}
		if err := s.store.CreateTenant(t); err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeJSON(w, http.StatusCreated, map[string]string{"status": "created", "id": t.ID})

	default:
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
	}
}

func (s *Server) handleTenantRoute(w http.ResponseWriter, r *http.Request) {
	subPath := strings.TrimPrefix(r.URL.Path, "/api/tenants/")
	if subPath == "" {
		writeError(w, http.StatusBadRequest, "Missing tenant ID")
		return
	}

	// Check if this is a statement request: /api/tenants/{id}/statement
	if strings.HasSuffix(subPath, "/statement") {
		tenantID := strings.TrimSuffix(subPath, "/statement")
		month := r.URL.Query().Get("month")
		stmt, err := s.store.GetTenantStatement(tenantID, month)
		if err != nil {
			writeError(w, http.StatusNotFound, err.Error())
			return
		}

		if r.URL.Query().Get("format") == "html" {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			renderStatementHTML(w, stmt)
			return
		}

		writeJSON(w, http.StatusOK, stmt)
		return
	}

	id := subPath
	switch r.Method {
	case http.MethodGet:
		tenant, err := s.store.GetTenant(id)
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		if tenant == nil {
			writeError(w, http.StatusNotFound, "Tenant not found")
			return
		}

		month := r.URL.Query().Get("month")
		daily, _ := s.store.GetDailyUsage(id, 30)
		features, _ := s.store.GetFeatureBreakdown(id, month)
		modelsList, _ := s.store.GetModelDistribution(id, month)
		logs, _ := s.store.GetRecentLogs(id, 50)

		writeJSON(w, http.StatusOK, map[string]any{
			"tenant":   tenant,
			"daily":    daily,
			"features": features,
			"models":   modelsList,
			"logs":     logs,
		})

	case http.MethodDelete:
		if err := s.store.DeleteTenant(id); err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"status": "deleted"})

	default:
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
	}
}

func (s *Server) handleFeatures(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}
	tenantID := r.URL.Query().Get("tenant_id")
	month := r.URL.Query().Get("month")
	features, err := s.store.GetFeatureBreakdown(tenantID, month)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, features)
}

func (s *Server) handleModels(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}
	tenantID := r.URL.Query().Get("tenant_id")
	month := r.URL.Query().Get("month")
	modelsList, err := s.store.GetModelDistribution(tenantID, month)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, modelsList)
}

func (s *Server) handleAlerts(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}
	var resolved *bool
	if resStr := r.URL.Query().Get("resolved"); resStr != "" {
		val := resStr == "true"
		resolved = &val
	}
	list, err := s.store.ListAlerts(resolved)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, list)
}

func (s *Server) handleAlertAction(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/alerts/")
	parts := strings.Split(path, "/")
	if len(parts) < 2 || parts[1] != "resolve" {
		writeError(w, http.StatusNotFound, "Endpoint not found")
		return
	}
	id := parts[0]
	if err := s.store.ResolveAlert(id); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "resolved"})
}

func (s *Server) handleRecommendations(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		var applied *bool
		if appStr := r.URL.Query().Get("applied"); appStr != "" {
			val := appStr == "true"
			applied = &val
		}
		list, err := s.store.ListRecommendations(applied)
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, list)

	case http.MethodPost:
		month := r.URL.Query().Get("month")
		recs, err := finops.GenerateRecommendations(s.store, month)
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{
			"status":    "generated",
			"count":     len(recs),
			"new_items": recs,
		})

	default:
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
	}
}

func (s *Server) handleRecommendationAction(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/recommendations/")
	parts := strings.Split(path, "/")
	if len(parts) < 2 || parts[1] != "apply" {
		writeError(w, http.StatusNotFound, "Endpoint not found")
		return
	}
	id := parts[0]
	if err := s.store.ApplyRecommendation(id); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "applied"})
}

func (s *Server) handleLogs(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}
	tenantID := r.URL.Query().Get("tenant_id")
	limit := 100
	if l := r.URL.Query().Get("limit"); l != "" {
		if val, err := strconv.Atoi(l); err == nil && val > 0 {
			limit = val
		}
	}
	logs, err := s.store.GetRecentLogs(tenantID, limit)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, logs)
}

func (s *Server) handleCacheAnalytics(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}
	month := r.URL.Query().Get("month")
	analytics, err := s.store.GetCacheAnalytics(month)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, analytics)
}

type SimulationRequest struct {
	SubscriptionIncreaseUSD float64           `json:"subscription_increase_usd"`
	ModelReplacements       map[string]string `json:"model_replacements"`
}

type SimulationResponse struct {
	OriginalRevenueUSD    float64 `json:"original_revenue_usd"`
	OriginalCostUSD       float64 `json:"original_cost_usd"`
	OriginalMarginUSD     float64 `json:"original_margin_usd"`
	OriginalMarginPct     float64 `json:"original_margin_pct"`
	OriginalUnprofitable  int     `json:"original_unprofitable"`

	SimulatedRevenueUSD   float64 `json:"simulated_revenue_usd"`
	SimulatedCostUSD      float64 `json:"simulated_cost_usd"`
	SimulatedMarginUSD    float64 `json:"simulated_margin_usd"`
	SimulatedMarginPct    float64 `json:"simulated_margin_pct"`
	SimulatedUnprofitable int     `json:"simulated_unprofitable"`

	MonthlyNetGainUSD     float64 `json:"monthly_net_gain_usd"`
}

func (s *Server) handleSimulator(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var req SimulationRequest
	_ = json.NewDecoder(r.Body).Decode(&req)

	month := r.URL.Query().Get("month")
	metricsList, err := s.store.GetTenantMetrics(month)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	var resp SimulationResponse

	for _, tm := range metricsList {
		resp.OriginalRevenueUSD += tm.MonthlyPlanUSD
		resp.OriginalCostUSD += tm.TotalCostUSD
		if tm.IsUnprofitable {
			resp.OriginalUnprofitable++
		}

		// Calculate simulated revenue
		simRev := tm.MonthlyPlanUSD
		if simRev > 0 {
			simRev += req.SubscriptionIncreaseUSD
		}
		resp.SimulatedRevenueUSD += simRev

		// Calculate simulated cost by checking tenant's feature model replacements
		simCost := tm.TotalCostUSD
		if len(req.ModelReplacements) > 0 {
			logs, _ := s.store.GetRecentLogs(tm.TenantID, 500)
			for _, l := range logs {
				if targetModel, ok := req.ModelReplacements[l.Model]; ok && targetModel != l.Model {
					origCost := l.CostUSD
					newCost := pricing.ComputeCost(targetModel, l.PromptTokens, l.CompletionTokens, l.CachedTokens)
					diff := origCost - newCost
					if diff > 0 {
						simCost -= diff
					}
				}
			}
		}
		if simCost < 0 {
			simCost = 0
		}
		resp.SimulatedCostUSD += simCost

		if simRev < simCost {
			resp.SimulatedUnprofitable++
		}
	}

	resp.OriginalMarginUSD = resp.OriginalRevenueUSD - resp.OriginalCostUSD
	if resp.OriginalRevenueUSD > 0 {
		resp.OriginalMarginPct = (resp.OriginalMarginUSD / resp.OriginalRevenueUSD) * 100.0
	}

	resp.SimulatedMarginUSD = resp.SimulatedRevenueUSD - resp.SimulatedCostUSD
	if resp.SimulatedRevenueUSD > 0 {
		resp.SimulatedMarginPct = (resp.SimulatedMarginUSD / resp.SimulatedRevenueUSD) * 100.0
	}

	resp.MonthlyNetGainUSD = resp.SimulatedMarginUSD - resp.OriginalMarginUSD

	writeJSON(w, http.StatusOK, resp)
}

func (s *Server) handleExportCSV(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}
	month := r.URL.Query().Get("month")
	metricsList, err := s.store.GetTenantMetrics(month)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", `attachment; filename="llmmeter_margins.csv"`)

	writer := csv.NewWriter(w)
	_ = writer.Write([]string{
		"tenant_id",
		"tenant_name",
		"monthly_plan_usd",
		"total_cost_usd",
		"gross_margin_usd",
		"gross_margin_pct",
		"total_tokens",
		"request_count",
		"is_unprofitable",
	})

	for _, m := range metricsList {
		_ = writer.Write([]string{
			m.TenantID,
			m.TenantName,
			fmt.Sprintf("%.2f", m.MonthlyPlanUSD),
			fmt.Sprintf("%.4f", m.TotalCostUSD),
			fmt.Sprintf("%.4f", m.GrossMarginUSD),
			fmt.Sprintf("%.2f%%", m.GrossMarginPct),
			strconv.FormatInt(m.TotalTokens, 10),
			strconv.FormatInt(m.RequestCount, 10),
			strconv.FormatBool(m.IsUnprofitable),
		})
	}
	writer.Flush()
}

func renderStatementHTML(w http.ResponseWriter, stmt *store.TenantStatement) {
	html := fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
<meta charset="utf-8">
<title>AI Usage &amp; Chargeback Statement — %s</title>
<style>
body { font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif; margin: 40px; color: #18181b; }
.header { border-bottom: 2px solid #e4e4e7; padding-bottom: 20px; margin-bottom: 30px; display: flex; justify-content: space-between; align-items: flex-start; }
.brand { font-size: 24px; font-weight: bold; font-family: monospace; color: #09090b; }
.badge { display: inline-block; padding: 4px 8px; border-radius: 4px; font-size: 11px; font-weight: 600; text-transform: uppercase; background: #e4e4e7; }
.meta { display: grid; grid-template-columns: repeat(3, 1fr); gap: 20px; margin-bottom: 30px; background: #f4f4f5; padding: 20px; border-radius: 8px; }
.meta-item span { display: block; font-size: 11px; color: #71717a; text-transform: uppercase; font-weight: 600; }
.meta-item strong { font-size: 18px; font-family: monospace; color: #09090b; }
table { width: 100%%; border-collapse: collapse; margin-bottom: 30px; font-size: 13px; }
th, td { text-align: left; padding: 10px 12px; border-bottom: 1px solid #e4e4e7; }
th { background: #f4f4f5; font-size: 11px; text-transform: uppercase; color: #71717a; font-family: monospace; }
td.mono { font-family: monospace; }
td.num { text-align: right; }
.summary { width: 350px; margin-left: auto; border-top: 2px solid #18181b; padding-top: 15px; }
.summary-row { display: flex; justify-content: space-between; padding: 6px 0; font-size: 14px; }
.summary-row.total { font-size: 16px; font-weight: bold; border-top: 1px solid #e4e4e7; margin-top: 6px; padding-top: 10px; }
.footer { margin-top: 50px; border-top: 1px solid #e4e4e7; padding-top: 20px; font-size: 11px; color: #a1a1aa; text-align: center; }
@media print { body { margin: 0; } }
</style>
</head>
<body>
<div class="header">
  <div>
    <div class="brand">llmmeter</div>
    <div style="font-size: 13px; color: #71717a; margin-top: 4px;">Tenant AI Usage &amp; Cost Chargeback Statement</div>
  </div>
  <div style="text-align: right;">
    <span class="badge">Billing Period: %s</span>
    <div style="font-size: 12px; color: #71717a; margin-top: 6px;">Generated: %s</div>
  </div>
</div>

<div class="meta">
  <div class="meta-item">
    <span>Customer Account</span>
    <strong>%s</strong>
    <div style="font-size: 11px; color: #71717a; font-family: monospace;">ID: %s</div>
  </div>
  <div class="meta-item">
    <span>Base Plan Tier</span>
    <strong>$%.2f/mo</strong>
    <div style="font-size: 11px; color: #71717a;">Flat Subscription</div>
  </div>
  <div class="meta-item">
    <span>Total AI Infrastructure</span>
    <strong>$%.2f</strong>
    <div style="font-size: 11px; color: #71717a;">%d Requests Metred</div>
  </div>
</div>

<h3>Itemized Route &amp; Feature Consumption</h3>
<table>
  <thead>
    <tr>
      <th>Feature Tag</th>
      <th>Foundation Model</th>
      <th class="num">Prompt Tok</th>
      <th class="num">Completion Tok</th>
      <th class="num">Cached Tok</th>
      <th class="num">Requests</th>
      <th class="num">Attributed Cost</th>
    </tr>
  </thead>
  <tbody>`,
		stmt.TenantName,
		stmt.Period,
		stmt.GeneratedAt.Format("2006-01-02 15:04"),
		stmt.TenantName,
		stmt.TenantID,
		stmt.MonthlyPlanUSD,
		stmt.TotalAIUsageUSD,
		stmt.TotalRequests,
	)

	for _, item := range stmt.LineItems {
		html += fmt.Sprintf(`
    <tr>
      <td class="mono"><strong>%s</strong></td>
      <td class="mono">%s</td>
      <td class="mono num">%d</td>
      <td class="mono num">%d</td>
      <td class="mono num">%d</td>
      <td class="mono num">%d</td>
      <td class="mono num"><strong>$%.4f</strong></td>
    </tr>`,
			item.FeatureTag,
			item.Model,
			item.PromptTokens,
			item.CompletionTokens,
			item.CachedTokens,
			item.RequestCount,
			item.TotalCostUSD,
		)
	}

	html += fmt.Sprintf(`
  </tbody>
</table>

<div class="summary">
  <div class="summary-row">
    <span>Subscription Base Revenue:</span>
    <span class="mono">$%.2f</span>
  </div>
  <div class="summary-row">
    <span>Total LLM Workload Cost:</span>
    <span class="mono">-$%.2f</span>
  </div>
  <div class="summary-row total">
    <span>Gross Margin:</span>
    <span class="mono">$%.2f (%.1f%%%%)</span>
  </div>
</div>

<div class="footer">
  This statement was automatically generated by llmmeter Multi-Provider AI FinOps Engine. Verified through cryptographic token metering.
</div>
</body>
</html>`,
		stmt.MonthlyPlanUSD,
		stmt.TotalAIUsageUSD,
		stmt.GrossMarginUSD,
		stmt.GrossMarginPct,
	)

	_, _ = w.Write([]byte(html))
}
