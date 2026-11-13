package store

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	_ "modernc.org/sqlite"
)

// Store wraps the SQLite database handle and FinOps queries.
type Store struct {
	db *sql.DB
}

// Tenant represents a SaaS tenant / customer.
type Tenant struct {
	ID               string    `json:"id"`
	Name             string    `json:"name"`
	MonthlyPlanUSD   float64   `json:"monthly_plan_usd"`
	MonthlyBudgetUSD *float64  `json:"monthly_budget_usd,omitempty"`
	WebhookURL       string    `json:"webhook_url,omitempty"`
	CreatedAt        time.Time `json:"created_at"`
}

// RequestLog represents a single LLM proxy execution.
type RequestLog struct {
	ID               string    `json:"id"`
	TenantID         string    `json:"tenant_id"`
	FeatureTag       string    `json:"feature_tag"`
	UserID           string    `json:"user_id,omitempty"`
	Provider         string    `json:"provider"`
	Model            string    `json:"model"`
	PromptTokens     int       `json:"prompt_tokens"`
	CompletionTokens int       `json:"completion_tokens"`
	CachedTokens     int       `json:"cached_tokens"`
	CostUSD          float64   `json:"cost_usd"`
	LatencyMS        int64     `json:"latency_ms"`
	Timestamp        time.Time `json:"timestamp"`
}

// TenantMetric represents monthly rollups and margin analytics.
type TenantMetric struct {
	TenantID         string   `json:"tenant_id"`
	TenantName       string   `json:"tenant_name"`
	MonthlyPlanUSD   float64  `json:"monthly_plan_usd"`
	MonthlyBudgetUSD *float64 `json:"monthly_budget_usd,omitempty"`
	TotalCostUSD     float64  `json:"total_cost_usd"`
	GrossMarginUSD   float64  `json:"gross_margin_usd"`
	GrossMarginPct   float64  `json:"gross_margin_pct"`
	TotalTokens      int64    `json:"total_tokens"`
	RequestCount     int64    `json:"request_count"`
	IsUnprofitable   bool     `json:"is_unprofitable"`
}

// FeatureMetric represents cost attributed to a specific feature tag.
type FeatureMetric struct {
	FeatureTag     string  `json:"feature_tag"`
	TotalCostUSD   float64 `json:"total_cost_usd"`
	TotalTokens    int64   `json:"total_tokens"`
	RequestCount   int64   `json:"request_count"`
	AvgCostUSD     float64 `json:"avg_cost_usd"`
	TopModel       string  `json:"top_model"`
}

// ModelMetric represents token volume and costs grouped by model.
type ModelMetric struct {
	Model        string  `json:"model"`
	Provider     string  `json:"provider"`
	TotalCostUSD float64 `json:"total_cost_usd"`
	TotalTokens  int64   `json:"total_tokens"`
	RequestCount int64   `json:"request_count"`
}

// DailyUsage represents daily spending over time.
type DailyUsage struct {
	Date         string  `json:"date"`
	TotalCostUSD float64 `json:"total_cost_usd"`
	TotalTokens  int64   `json:"total_tokens"`
	RequestCount int64   `json:"request_count"`
}

// Alert represents an anomaly, budget breach, or negative margin alarm.
type Alert struct {
	ID        string    `json:"id"`
	TenantID  string    `json:"tenant_id"`
	Type      string    `json:"type"` // anomaly, budget, negative_margin
	Message   string    `json:"message"`
	Severity  string    `json:"severity"` // info, warning, critical
	Resolved  bool      `json:"resolved"`
	Timestamp time.Time `json:"timestamp"`
}

// CacheFeatureMetric represents caching stats for a specific feature.
type CacheFeatureMetric struct {
	FeatureTag   string  `json:"feature_tag"`
	PromptTokens int64   `json:"prompt_tokens"`
	CachedTokens int64   `json:"cached_tokens"`
	HitRatePct   float64 `json:"hit_rate_pct"`
	SavedUSD     float64 `json:"saved_usd"`
}

// CacheAnalytics represents global cache efficiency and dollars saved.
type CacheAnalytics struct {
	TotalPromptTokens int64                `json:"total_prompt_tokens"`
	TotalCachedTokens int64                `json:"total_cached_tokens"`
	OverallHitRatePct float64              `json:"overall_hit_rate_pct"`
	TotalSavedUSD     float64              `json:"total_saved_usd"`
	TopFeatures       []CacheFeatureMetric `json:"top_features"`
}

// StatementLineItem itemizes usage on a customer invoice statement.
type StatementLineItem struct {
	FeatureTag       string  `json:"feature_tag"`
	Model            string  `json:"model"`
	PromptTokens     int64   `json:"prompt_tokens"`
	CompletionTokens int64   `json:"completion_tokens"`
	CachedTokens     int64   `json:"cached_tokens"`
	RequestCount     int64   `json:"request_count"`
	TotalCostUSD     float64 `json:"total_cost_usd"`
}

// TenantStatement represents a formal billing and chargeback statement.
type TenantStatement struct {
	TenantID         string              `json:"tenant_id"`
	TenantName       string              `json:"tenant_name"`
	Period           string              `json:"period"`
	MonthlyPlanUSD   float64             `json:"monthly_plan_usd"`
	MonthlyBudgetUSD *float64            `json:"monthly_budget_usd,omitempty"`
	TotalAIUsageUSD  float64             `json:"total_ai_usage_usd"`
	GrossMarginUSD   float64             `json:"gross_margin_usd"`
	GrossMarginPct   float64             `json:"gross_margin_pct"`
	TotalTokens      int64               `json:"total_tokens"`
	TotalRequests    int64               `json:"total_requests"`
	LineItems        []StatementLineItem `json:"line_items"`
	GeneratedAt      time.Time           `json:"generated_at"`
}

// Recommendation represents an automatic downgrade / optimization proposal.
type Recommendation struct {
	ID                string    `json:"id"`
	TenantID          string    `json:"tenant_id"`
	FeatureTag        string    `json:"feature_tag"`
	CurrentModel      string    `json:"current_model"`
	SuggestedModel    string    `json:"suggested_model"`
	MonthlySavingsUSD float64   `json:"monthly_savings_usd"`
	Reason            string    `json:"reason"`
	Applied           bool      `json:"applied"`
	Timestamp         time.Time `json:"timestamp"`
}

// OverviewSummary provides global KPIs for the FinOps Studio.
type OverviewSummary struct {
	TotalRevenueUSD    float64 `json:"total_revenue_usd"`
	TotalCostUSD       float64 `json:"total_cost_usd"`
	NetMarginUSD       float64 `json:"net_margin_usd"`
	NetMarginPct       float64 `json:"net_margin_pct"`
	ActiveTenantsCount int     `json:"active_tenants_count"`
	UnprofitableCount  int     `json:"unprofitable_count"`
	TotalTokensMonth   int64   `json:"total_tokens_month"`
	ActiveAlertsCount  int     `json:"active_alerts_count"`
}

// New initializes SQLite with WAL mode and runs migrations.
func New(dbPath string) (*Store, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}

	// Optimize for high concurrency and zero lock delays
	pragmas := []string{
		"PRAGMA journal_mode = WAL;",
		"PRAGMA busy_timeout = 5000;",
		"PRAGMA synchronous = NORMAL;",
		"PRAGMA foreign_keys = ON;",
	}
	for _, p := range pragmas {
		if _, err := db.Exec(p); err != nil {
			return nil, fmt.Errorf("exec pragma %s: %w", p, err)
		}
	}

	s := &Store{db: db}
	if err := s.migrate(); err != nil {
		return nil, fmt.Errorf("migrate: %w", err)
	}

	return s, nil
}

func (s *Store) Close() error {
	return s.db.Close()
}

func (s *Store) migrate() error {
	schema := `
	CREATE TABLE IF NOT EXISTS tenants (
		id TEXT PRIMARY KEY,
		name TEXT NOT NULL,
		monthly_plan_usd REAL NOT NULL DEFAULT 0.0,
		monthly_budget_usd REAL,
		webhook_url TEXT,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS request_logs (
		id TEXT PRIMARY KEY,
		tenant_id TEXT NOT NULL REFERENCES tenants(id),
		feature_tag TEXT NOT NULL,
		user_id TEXT,
		provider TEXT NOT NULL,
		model TEXT NOT NULL,
		prompt_tokens INTEGER NOT NULL,
		completion_tokens INTEGER NOT NULL,
		cached_tokens INTEGER DEFAULT 0,
		cost_usd REAL NOT NULL,
		latency_ms INTEGER NOT NULL,
		timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);

	CREATE INDEX IF NOT EXISTS idx_req_tenant_time ON request_logs(tenant_id, timestamp DESC);
	CREATE INDEX IF NOT EXISTS idx_req_feature ON request_logs(feature_tag, timestamp DESC);
	CREATE INDEX IF NOT EXISTS idx_req_model ON request_logs(model, timestamp DESC);
	CREATE INDEX IF NOT EXISTS idx_req_timestamp ON request_logs(timestamp DESC);

	CREATE TABLE IF NOT EXISTS alerts (
		id TEXT PRIMARY KEY,
		tenant_id TEXT NOT NULL REFERENCES tenants(id),
		type TEXT NOT NULL,
		message TEXT NOT NULL,
		severity TEXT NOT NULL,
		resolved BOOLEAN NOT NULL DEFAULT 0,
		timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);
	CREATE INDEX IF NOT EXISTS idx_alerts_resolved ON alerts(resolved, timestamp DESC);

	CREATE TABLE IF NOT EXISTS recommendations (
		id TEXT PRIMARY KEY,
		tenant_id TEXT NOT NULL REFERENCES tenants(id),
		feature_tag TEXT NOT NULL,
		current_model TEXT NOT NULL,
		suggested_model TEXT NOT NULL,
		monthly_savings_usd REAL NOT NULL,
		reason TEXT NOT NULL,
		applied BOOLEAN NOT NULL DEFAULT 0,
		timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);
	CREATE INDEX IF NOT EXISTS idx_recs_applied ON recommendations(applied, timestamp DESC);
	`
	_, err := s.db.Exec(schema)
	return err
}

// --- Tenant Management ---

func (s *Store) CreateTenant(t Tenant) error {
	if t.ID == "" {
		t.ID = strings.ToLower(strings.ReplaceAll(t.Name, " ", "-"))
	}
	query := `
	INSERT INTO tenants (id, name, monthly_plan_usd, monthly_budget_usd, webhook_url, created_at)
	VALUES (?, ?, ?, ?, ?, COALESCE(NULLIF(?, ''), CURRENT_TIMESTAMP))
	ON CONFLICT(id) DO UPDATE SET
		name = excluded.name,
		monthly_plan_usd = excluded.monthly_plan_usd,
		monthly_budget_usd = excluded.monthly_budget_usd,
		webhook_url = excluded.webhook_url
	`
	var createdAtStr string
	if !t.CreatedAt.IsZero() {
		createdAtStr = t.CreatedAt.Format("2006-01-02 15:04:05")
	}
	_, err := s.db.Exec(query, t.ID, t.Name, t.MonthlyPlanUSD, t.MonthlyBudgetUSD, t.WebhookURL, createdAtStr)
	return err
}

func (s *Store) GetTenant(id string) (*Tenant, error) {
	row := s.db.QueryRow("SELECT id, name, monthly_plan_usd, monthly_budget_usd, webhook_url, created_at FROM tenants WHERE id = ?", id)
	var t Tenant
	var createdStr string
	var webhook sql.NullString
	var budget sql.NullFloat64

	err := row.Scan(&t.ID, &t.Name, &t.MonthlyPlanUSD, &budget, &webhook, &createdStr)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if budget.Valid {
		t.MonthlyBudgetUSD = &budget.Float64
	}
	if webhook.Valid {
		t.WebhookURL = webhook.String
	}
	t.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", createdStr)
	return &t, nil
}

func (s *Store) ListTenants() ([]Tenant, error) {
	rows, err := s.db.Query("SELECT id, name, monthly_plan_usd, monthly_budget_usd, webhook_url, created_at FROM tenants ORDER BY name ASC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []Tenant
	for rows.Next() {
		var t Tenant
		var createdStr string
		var webhook sql.NullString
		var budget sql.NullFloat64
		if err := rows.Scan(&t.ID, &t.Name, &t.MonthlyPlanUSD, &budget, &webhook, &createdStr); err != nil {
			return nil, err
		}
		if budget.Valid {
			t.MonthlyBudgetUSD = &budget.Float64
		}
		if webhook.Valid {
			t.WebhookURL = webhook.String
		}
		t.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", createdStr)
		list = append(list, t)
	}
	return list, nil
}

func (s *Store) DeleteTenant(id string) error {
	_, err := s.db.Exec("DELETE FROM tenants WHERE id = ?", id)
	return err
}

// --- Request Logging ---

func (s *Store) LogRequest(r RequestLog) error {
	if r.ID == "" {
		r.ID = uuid.New().String()
	}
	query := `
	INSERT INTO request_logs (id, tenant_id, feature_tag, user_id, provider, model, prompt_tokens, completion_tokens, cached_tokens, cost_usd, latency_ms, timestamp)
	VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, COALESCE(NULLIF(?, ''), CURRENT_TIMESTAMP))
	`
	var timeStr string
	if !r.Timestamp.IsZero() {
		timeStr = r.Timestamp.Format("2006-01-02 15:04:05")
	}
	_, err := s.db.Exec(query, r.ID, r.TenantID, r.FeatureTag, r.UserID, r.Provider, r.Model, r.PromptTokens, r.CompletionTokens, r.CachedTokens, r.CostUSD, r.LatencyMS, timeStr)
	return err
}

func (s *Store) GetRecentLogs(tenantID string, limit int) ([]RequestLog, error) {
	if limit <= 0 {
		limit = 100
	}
	var rows *sql.Rows
	var err error
	if tenantID != "" {
		rows, err = s.db.Query(`
			SELECT id, tenant_id, feature_tag, COALESCE(user_id, ''), provider, model, prompt_tokens, completion_tokens, cached_tokens, cost_usd, latency_ms, timestamp
			FROM request_logs
			WHERE tenant_id = ?
			ORDER BY timestamp DESC
			LIMIT ?`, tenantID, limit)
	} else {
		rows, err = s.db.Query(`
			SELECT id, tenant_id, feature_tag, COALESCE(user_id, ''), provider, model, prompt_tokens, completion_tokens, cached_tokens, cost_usd, latency_ms, timestamp
			FROM request_logs
			ORDER BY timestamp DESC
			LIMIT ?`, limit)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []RequestLog
	for rows.Next() {
		var l RequestLog
		var timeStr string
		if err := rows.Scan(&l.ID, &l.TenantID, &l.FeatureTag, &l.UserID, &l.Provider, &l.Model, &l.PromptTokens, &l.CompletionTokens, &l.CachedTokens, &l.CostUSD, &l.LatencyMS, &timeStr); err != nil {
			return nil, err
		}
		l.Timestamp, _ = time.Parse("2006-01-02 15:04:05", timeStr)
		logs = append(logs, l)
	}
	return logs, nil
}

// --- Margin Analytics & Unit Economics ---

// GetTenantMetrics computes monthly gross margin per tenant.
func (s *Store) GetTenantMetrics(month string) ([]TenantMetric, error) {
	if month == "" {
		month = time.Now().Format("2006-01")
	}
	startTime := month + "-01 00:00:00"

	query := `
	SELECT 
		t.id, 
		t.name, 
		t.monthly_plan_usd, 
		t.monthly_budget_usd,
		COALESCE(SUM(r.cost_usd), 0.0) AS total_cost_usd,
		COALESCE(SUM(r.prompt_tokens + r.completion_tokens), 0) AS total_tokens,
		COUNT(r.id) AS request_count
	FROM tenants t
	LEFT JOIN request_logs r ON t.id = r.tenant_id AND r.timestamp >= ?
	GROUP BY t.id
	ORDER BY total_cost_usd DESC
	`
	rows, err := s.db.Query(query, startTime)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var metrics []TenantMetric
	for rows.Next() {
		var m TenantMetric
		var budget sql.NullFloat64
		if err := rows.Scan(&m.TenantID, &m.TenantName, &m.MonthlyPlanUSD, &budget, &m.TotalCostUSD, &m.TotalTokens, &m.RequestCount); err != nil {
			return nil, err
		}
		if budget.Valid {
			m.MonthlyBudgetUSD = &budget.Float64
		}

		m.GrossMarginUSD = m.MonthlyPlanUSD - m.TotalCostUSD
		if m.MonthlyPlanUSD > 0 {
			m.GrossMarginPct = (m.GrossMarginUSD / m.MonthlyPlanUSD) * 100.0
		} else {
			m.GrossMarginPct = 0.0
		}
		m.IsUnprofitable = m.GrossMarginUSD < 0

		metrics = append(metrics, m)
	}
	return metrics, nil
}

// GetUnprofitableTenants returns only accounts with negative gross margin.
func (s *Store) GetUnprofitableTenants(month string) ([]TenantMetric, error) {
	all, err := s.GetTenantMetrics(month)
	if err != nil {
		return nil, err
	}
	var unprofitable []TenantMetric
	for _, m := range all {
		if m.IsUnprofitable {
			unprofitable = append(unprofitable, m)
		}
	}
	return unprofitable, nil
}

// GetFeatureBreakdown ranks features by spend.
func (s *Store) GetFeatureBreakdown(tenantID, month string) ([]FeatureMetric, error) {
	if month == "" {
		month = time.Now().Format("2006-01")
	}
	startTime := month + "-01 00:00:00"

	where := "timestamp >= ?"
	args := []interface{}{startTime}
	if tenantID != "" {
		where += " AND tenant_id = ?"
		args = append(args, tenantID)
	}

	query := fmt.Sprintf(`
	SELECT 
		feature_tag, 
		SUM(cost_usd) as total_cost, 
		SUM(prompt_tokens + completion_tokens) as total_tokens, 
		COUNT(id) as req_count,
		AVG(cost_usd) as avg_cost,
		(SELECT model FROM request_logs r2 WHERE r2.feature_tag = r1.feature_tag GROUP BY model ORDER BY COUNT(*) DESC LIMIT 1) as top_model
	FROM request_logs r1
	WHERE %s
	GROUP BY feature_tag
	ORDER BY total_cost DESC
	`, where)

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var res []FeatureMetric
	for rows.Next() {
		var f FeatureMetric
		var topModel sql.NullString
		if err := rows.Scan(&f.FeatureTag, &f.TotalCostUSD, &f.TotalTokens, &f.RequestCount, &f.AvgCostUSD, &topModel); err != nil {
			return nil, err
		}
		if topModel.Valid {
			f.TopModel = topModel.String
		}
		res = append(res, f)
	}
	return res, nil
}

// GetModelDistribution aggregates volume by model.
func (s *Store) GetModelDistribution(tenantID, month string) ([]ModelMetric, error) {
	if month == "" {
		month = time.Now().Format("2006-01")
	}
	startTime := month + "-01 00:00:00"

	where := "timestamp >= ?"
	args := []interface{}{startTime}
	if tenantID != "" {
		where += " AND tenant_id = ?"
		args = append(args, tenantID)
	}

	query := fmt.Sprintf(`
	SELECT model, provider, SUM(cost_usd), SUM(prompt_tokens + completion_tokens), COUNT(id)
	FROM request_logs
	WHERE %s
	GROUP BY model, provider
	ORDER BY SUM(cost_usd) DESC
	`, where)

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var res []ModelMetric
	for rows.Next() {
		var m ModelMetric
		if err := rows.Scan(&m.Model, &m.Provider, &m.TotalCostUSD, &m.TotalTokens, &m.RequestCount); err != nil {
			return nil, err
		}
		res = append(res, m)
	}
	return res, nil
}

// GetDailyUsage returns daily time-series spending.
func (s *Store) GetDailyUsage(tenantID string, days int) ([]DailyUsage, error) {
	if days <= 0 {
		days = 30
	}
	cutoff := time.Now().AddDate(0, 0, -days).Format("2006-01-02 15:04:05")

	where := "timestamp >= ?"
	args := []interface{}{cutoff}
	if tenantID != "" {
		where += " AND tenant_id = ?"
		args = append(args, tenantID)
	}

	query := fmt.Sprintf(`
	SELECT substr(timestamp, 1, 10) as date_val, SUM(cost_usd), SUM(prompt_tokens + completion_tokens), COUNT(id)
	FROM request_logs
	WHERE %s
	GROUP BY date_val
	ORDER BY date_val ASC
	`, where)

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var res []DailyUsage
	for rows.Next() {
		var d DailyUsage
		if err := rows.Scan(&d.Date, &d.TotalCostUSD, &d.TotalTokens, &d.RequestCount); err != nil {
			return nil, err
		}
		res = append(res, d)
	}
	return res, nil
}

// GetRollingAverageCost computes average request cost over the past N days.
func (s *Store) GetRollingAverageCost(tenantID string, days int) (float64, int, error) {
	cutoff := time.Now().AddDate(0, 0, -days).Format("2006-01-02 15:04:05")
	row := s.db.QueryRow(`
		SELECT COALESCE(AVG(cost_usd), 0.0), COUNT(id)
		FROM request_logs
		WHERE tenant_id = ? AND timestamp >= ?`, tenantID, cutoff)

	var avg float64
	var count int
	err := row.Scan(&avg, &count)
	return avg, count, err
}

// GetTenantMonthSpend returns current month spend for a tenant.
func (s *Store) GetTenantMonthSpend(tenantID string) (float64, error) {
	monthStart := time.Now().Format("2006-01") + "-01 00:00:00"
	row := s.db.QueryRow(`SELECT COALESCE(SUM(cost_usd), 0.0) FROM request_logs WHERE tenant_id = ? AND timestamp >= ?`, tenantID, monthStart)
	var sum float64
	err := row.Scan(&sum)
	return sum, err
}

// GetOverviewSummary returns aggregate figures for the dashboard header.
func (s *Store) GetOverviewSummary(month string) (OverviewSummary, error) {
	metrics, err := s.GetTenantMetrics(month)
	if err != nil {
		return OverviewSummary{}, err
	}

	var sum OverviewSummary
	sum.ActiveTenantsCount = len(metrics)

	for _, m := range metrics {
		sum.TotalRevenueUSD += m.MonthlyPlanUSD
		sum.TotalCostUSD += m.TotalCostUSD
		sum.TotalTokensMonth += m.TotalTokens
		if m.IsUnprofitable {
			sum.UnprofitableCount++
		}
	}

	sum.NetMarginUSD = sum.TotalRevenueUSD - sum.TotalCostUSD
	if sum.TotalRevenueUSD > 0 {
		sum.NetMarginPct = (sum.NetMarginUSD / sum.TotalRevenueUSD) * 100.0
	}

	// Active alerts count
	row := s.db.QueryRow("SELECT COUNT(*) FROM alerts WHERE resolved = 0")
	_ = row.Scan(&sum.ActiveAlertsCount)

	return sum, nil
}

// --- Alerts ---

func (s *Store) CreateAlert(a Alert) error {
	if a.ID == "" {
		a.ID = uuid.New().String()
	}
	query := `
	INSERT INTO alerts (id, tenant_id, type, message, severity, resolved, timestamp)
	VALUES (?, ?, ?, ?, ?, ?, COALESCE(NULLIF(?, ''), CURRENT_TIMESTAMP))
	`
	var timeStr string
	if !a.Timestamp.IsZero() {
		timeStr = a.Timestamp.Format("2006-01-02 15:04:05")
	}
	_, err := s.db.Exec(query, a.ID, a.TenantID, a.Type, a.Message, a.Severity, a.Resolved, timeStr)
	return err
}

func (s *Store) ListAlerts(resolved *bool) ([]Alert, error) {
	query := "SELECT id, tenant_id, type, message, severity, resolved, timestamp FROM alerts"
	var args []interface{}
	if resolved != nil {
		query += " WHERE resolved = ?"
		args = append(args, *resolved)
	}
	query += " ORDER BY timestamp DESC LIMIT 200"

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []Alert
	for rows.Next() {
		var a Alert
		var timeStr string
		if err := rows.Scan(&a.ID, &a.TenantID, &a.Type, &a.Message, &a.Severity, &a.Resolved, &timeStr); err != nil {
			return nil, err
		}
		a.Timestamp, _ = time.Parse("2006-01-02 15:04:05", timeStr)
		list = append(list, a)
	}
	return list, nil
}

func (s *Store) ResolveAlert(id string) error {
	_, err := s.db.Exec("UPDATE alerts SET resolved = 1 WHERE id = ?", id)
	return err
}

// --- Recommendations ---

func (s *Store) CreateRecommendation(r Recommendation) error {
	if r.ID == "" {
		r.ID = uuid.New().String()
	}
	query := `
	INSERT INTO recommendations (id, tenant_id, feature_tag, current_model, suggested_model, monthly_savings_usd, reason, applied, timestamp)
	VALUES (?, ?, ?, ?, ?, ?, ?, ?, COALESCE(NULLIF(?, ''), CURRENT_TIMESTAMP))
	`
	var timeStr string
	if !r.Timestamp.IsZero() {
		timeStr = r.Timestamp.Format("2006-01-02 15:04:05")
	}
	_, err := s.db.Exec(query, r.ID, r.TenantID, r.FeatureTag, r.CurrentModel, r.SuggestedModel, r.MonthlySavingsUSD, r.Reason, r.Applied, timeStr)
	return err
}

func (s *Store) ListRecommendations(applied *bool) ([]Recommendation, error) {
	query := "SELECT id, tenant_id, feature_tag, current_model, suggested_model, monthly_savings_usd, reason, applied, timestamp FROM recommendations"
	var args []interface{}
	if applied != nil {
		query += " WHERE applied = ?"
		args = append(args, *applied)
	}
	query += " ORDER BY monthly_savings_usd DESC LIMIT 100"

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []Recommendation
	for rows.Next() {
		var r Recommendation
		var timeStr string
		if err := rows.Scan(&r.ID, &r.TenantID, &r.FeatureTag, &r.CurrentModel, &r.SuggestedModel, &r.MonthlySavingsUSD, &r.Reason, &r.Applied, &timeStr); err != nil {
			return nil, err
		}
		r.Timestamp, _ = time.Parse("2006-01-02 15:04:05", timeStr)
		list = append(list, r)
	}
	return list, nil
}

func (s *Store) ApplyRecommendation(id string) error {
	_, err := s.db.Exec("UPDATE recommendations SET applied = 1 WHERE id = ?", id)
	return err
}

// GetCacheAnalytics computes prompt cache efficiency and saved dollars.
func (s *Store) GetCacheAnalytics(month string) (CacheAnalytics, error) {
	if month == "" {
		month = time.Now().Format("2006-01")
	}
	startTime := month + "-01 00:00:00"

	query := `
	SELECT 
		feature_tag,
		SUM(prompt_tokens) as total_prompt,
		SUM(cached_tokens) as total_cached
	FROM request_logs
	WHERE timestamp >= ?
	GROUP BY feature_tag
	ORDER BY total_cached DESC
	`
	rows, err := s.db.Query(query, startTime)
	if err != nil {
		return CacheAnalytics{}, err
	}
	defer rows.Close()

	var totalPrompt, totalCached int64
	var features []CacheFeatureMetric

	for rows.Next() {
		var f CacheFeatureMetric
		if err := rows.Scan(&f.FeatureTag, &f.PromptTokens, &f.CachedTokens); err != nil {
			return CacheAnalytics{}, err
		}
		totalPrompt += f.PromptTokens
		totalCached += f.CachedTokens

		if f.PromptTokens > 0 {
			f.HitRatePct = (float64(f.CachedTokens) / float64(f.PromptTokens)) * 100.0
		}
		// Average prompt cache savings is ~$1.75 per 1M cached tokens across Claude/GPT-4o
		f.SavedUSD = (float64(f.CachedTokens) / 1_000_000.0) * 1.75
		features = append(features, f)
	}

	overallRate := 0.0
	if totalPrompt > 0 {
		overallRate = (float64(totalCached) / float64(totalPrompt)) * 100.0
	}
	totalSavedUSD := (float64(totalCached) / 1_000_000.0) * 1.75

	return CacheAnalytics{
		TotalPromptTokens: totalPrompt,
		TotalCachedTokens: totalCached,
		OverallHitRatePct: overallRate,
		TotalSavedUSD:     totalSavedUSD,
		TopFeatures:       features,
	}, nil
}

// GetTenantStatement compiles a full customer invoice statement with itemized route billing.
func (s *Store) GetTenantStatement(tenantID, month string) (*TenantStatement, error) {
	if month == "" {
		month = time.Now().Format("2006-01")
	}
	startTime := month + "-01 00:00:00"

	tenant, err := s.GetTenant(tenantID)
	if err != nil {
		return nil, err
	}
	if tenant == nil {
		return nil, fmt.Errorf("tenant %s not found", tenantID)
	}

	query := `
	SELECT 
		feature_tag,
		model,
		SUM(prompt_tokens) as prompt_tok,
		SUM(completion_tokens) as comp_tok,
		SUM(cached_tokens) as cached_tok,
		COUNT(id) as req_count,
		SUM(cost_usd) as cost
	FROM request_logs
	WHERE tenant_id = ? AND timestamp >= ?
	GROUP BY feature_tag, model
	ORDER BY cost DESC
	`
	rows, err := s.db.Query(query, tenantID, startTime)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	stmt := &TenantStatement{
		TenantID:         tenant.ID,
		TenantName:       tenant.Name,
		Period:           month,
		MonthlyPlanUSD:   tenant.MonthlyPlanUSD,
		MonthlyBudgetUSD: tenant.MonthlyBudgetUSD,
		GeneratedAt:      time.Now(),
	}

	for rows.Next() {
		var item StatementLineItem
		if err := rows.Scan(&item.FeatureTag, &item.Model, &item.PromptTokens, &item.CompletionTokens, &item.CachedTokens, &item.RequestCount, &item.TotalCostUSD); err != nil {
			return nil, err
		}
		stmt.TotalAIUsageUSD += item.TotalCostUSD
		stmt.TotalTokens += (item.PromptTokens + item.CompletionTokens)
		stmt.TotalRequests += item.RequestCount
		stmt.LineItems = append(stmt.LineItems, item)
	}

	stmt.GrossMarginUSD = stmt.MonthlyPlanUSD - stmt.TotalAIUsageUSD
	if stmt.MonthlyPlanUSD > 0 {
		stmt.GrossMarginPct = (stmt.GrossMarginUSD / stmt.MonthlyPlanUSD) * 100.0
	}

	return stmt, nil
}
