package api

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/alexandrmotologa/llmmeter/internal/finops"
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
	mux.HandleFunc("/api/tenants/", s.handleTenantDetail)
	mux.HandleFunc("/api/features", s.handleFeatures)
	mux.HandleFunc("/api/models", s.handleModels)
	mux.HandleFunc("/api/alerts", s.handleAlerts)
	mux.HandleFunc("/api/alerts/", s.handleAlertAction)
	mux.HandleFunc("/api/recommendations", s.handleRecommendations)
	mux.HandleFunc("/api/recommendations/", s.handleRecommendationAction)
	mux.HandleFunc("/api/logs", s.handleLogs)
	mux.HandleFunc("/api/export/csv", s.handleExportCSV)
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

		metrics, err := s.store.GetTenantMetrics(month)
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, metrics)

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

func (s *Server) handleTenantDetail(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/api/tenants/")
	if id == "" {
		writeError(w, http.StatusBadRequest, "Missing tenant ID")
		return
	}

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
		models, _ := s.store.GetModelDistribution(id, month)
		logs, _ := s.store.GetRecentLogs(id, 50)

		writeJSON(w, http.StatusOK, map[string]any{
			"tenant":    tenant,
			"daily":     daily,
			"features":  features,
			"models":    models,
			"logs":      logs,
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
	models, err := s.store.GetModelDistribution(tenantID, month)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, models)
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
		// Trigger recommendation generation
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

func (s *Server) handleExportCSV(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}
	month := r.URL.Query().Get("month")
	metrics, err := s.store.GetTenantMetrics(month)
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

	for _, m := range metrics {
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
