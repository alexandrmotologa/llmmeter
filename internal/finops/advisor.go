package finops

import (
	"fmt"
	"strings"
	"time"

	"github.com/alexandrmotologa/llmmeter/internal/alerts"
	"github.com/alexandrmotologa/llmmeter/internal/pricing"
	"github.com/alexandrmotologa/llmmeter/internal/store"
)

// DowngradeOption maps an expensive model to a cost-effective alternative.
type DowngradeOption struct {
	SuggestedModel string
	SavingsRatio   float64 // e.g. 0.70 means 70% cost reduction
	ReasonTemplate string
}

var downgradeCatalog = map[string]DowngradeOption{
	"claude-3-5-sonnet-20241022": {
		SuggestedModel: "claude-3-5-haiku-20241022",
		SavingsRatio:   0.70,
		ReasonTemplate: "Claude 3.5 Haiku delivers near-Sonnet reasoning at 70% lower token cost for high-frequency routes.",
	},
	"claude-3-5-sonnet": {
		SuggestedModel: "claude-3-5-haiku",
		SavingsRatio:   0.70,
		ReasonTemplate: "Claude 3.5 Haiku delivers near-Sonnet reasoning at 70% lower token cost for high-frequency routes.",
	},
	"claude-3-opus-20240229": {
		SuggestedModel: "claude-3-5-sonnet-20241022",
		SavingsRatio:   0.80,
		ReasonTemplate: "Claude 3.5 Sonnet outperforms Opus on 95% of software and extraction tasks at 80% lower cost.",
	},
	"claude-3-opus": {
		SuggestedModel: "claude-3-5-sonnet",
		SavingsRatio:   0.80,
		ReasonTemplate: "Claude 3.5 Sonnet outperforms Opus on 95% of software and extraction tasks at 80% lower cost.",
	},
	"gpt-4o": {
		SuggestedModel: "gpt-4o-mini",
		SavingsRatio:   0.94,
		ReasonTemplate: "GPT-4o-mini delivers fast responses with 94% cost reduction, ideal for extraction, classification, and summarization.",
	},
	"gpt-4-turbo": {
		SuggestedModel: "gpt-4o-mini",
		SavingsRatio:   0.98,
		ReasonTemplate: "Migrate legacy GPT-4 Turbo routes to GPT-4o-mini to reduce monthly LLM spend by up to 98%.",
	},
	"gpt-4": {
		SuggestedModel: "gpt-4o",
		SavingsRatio:   0.88,
		ReasonTemplate: "GPT-4o is significantly faster and 88% cheaper than legacy GPT-4.",
	},
	"o1": {
		SuggestedModel: "o3-mini",
		SavingsRatio:   0.92,
		ReasonTemplate: "o3-mini provides specialized reasoning with customizable effort at 92% lower cost than full o1.",
	},
	"gemini-1.5-pro": {
		SuggestedModel: "gemini-1.5-flash",
		SavingsRatio:   0.94,
		ReasonTemplate: "Gemini 1.5 Flash provides sub-second latency and 94% cost reduction with 1M context support.",
	},
	"llama-3.1-70b-versatile": {
		SuggestedModel: "llama-3.1-8b-instant",
		SavingsRatio:   0.90,
		ReasonTemplate: "Llama 3.1 8B provides 90% cost savings for lightweight conversational and routing agents.",
	},
}

// CheckRequestAnomaly evaluates if a single request significantly exceeds the tenant's historical baseline.
func CheckRequestAnomaly(s *store.Store, tenantID, featureTag, model string, costUSD float64) (*store.Alert, error) {
	if costUSD < 0.10 {
		return nil, nil // Ignore micro-cent requests
	}

	avgCost, count, err := s.GetRollingAverageCost(tenantID, 30)
	if err != nil {
		return nil, err
	}

	if count >= 5 && avgCost > 0 && costUSD > (avgCost*3.0) {
		message := fmt.Sprintf(
			"Spike detected: Request on %s (%s) cost $%.4f, which is %.1fx above 30-day average ($%.4f)",
			featureTag, model, costUSD, costUSD/avgCost, avgCost,
		)
		alert := store.Alert{
			TenantID:  tenantID,
			Type:      "anomaly",
			Message:   message,
			Severity:  "warning",
			Timestamp: time.Now(),
		}
		_ = s.CreateAlert(alert)

		// Check webhook
		if t, err := s.GetTenant(tenantID); err == nil && t != nil && t.WebhookURL != "" {
			alerts.Dispatch(t.WebhookURL, alerts.WebhookPayload{
				Event:     "anomaly_detected",
				TenantID:  tenantID,
				Type:      "anomaly",
				Severity:  "warning",
				Message:   message,
				SpendUSD:  costUSD,
				Timestamp: time.Now(),
			})
		}
		return &alert, nil
	}
	return nil, nil
}

// CheckBudgetCap verifies whether the tenant has exceeded their monthly budget cap.
func CheckBudgetCap(s *store.Store, tenant *store.Tenant, additionalCost float64) (bool, float64, *store.Alert, error) {
	if tenant == nil || tenant.MonthlyBudgetUSD == nil || *tenant.MonthlyBudgetUSD <= 0 {
		return false, 0, nil, nil
	}

	currentSpend, err := s.GetTenantMonthSpend(tenant.ID)
	if err != nil {
		return false, 0, nil, err
	}

	budget := *tenant.MonthlyBudgetUSD
	projected := currentSpend + additionalCost

	if projected >= budget {
		message := fmt.Sprintf(
			"Monthly budget cap reached: Tenant has spent $%.2f of $%.2f monthly budget cap.",
			projected, budget,
		)
		alert := store.Alert{
			TenantID:  tenant.ID,
			Type:      "budget",
			Message:   message,
			Severity:  "critical",
			Timestamp: time.Now(),
		}
		_ = s.CreateAlert(alert)

		if tenant.WebhookURL != "" {
			alerts.Dispatch(tenant.WebhookURL, alerts.WebhookPayload{
				Event:     "budget_exceeded",
				TenantID:  tenant.ID,
				Type:      "budget",
				Severity:  "critical",
				Message:   message,
				SpendUSD:  projected,
				BudgetUSD: budget,
				Timestamp: time.Now(),
			})
		}
		return true, projected, &alert, nil
	}

	return false, projected, nil, nil
}

// CheckNegativeMargin alerts when a customer's LLM consumption exceeds their subscription payment.
func CheckNegativeMargin(s *store.Store, tenant *store.Tenant) (*store.Alert, error) {
	if tenant == nil || tenant.MonthlyPlanUSD <= 0 {
		return nil, nil
	}

	currentSpend, err := s.GetTenantMonthSpend(tenant.ID)
	if err != nil {
		return nil, err
	}

	if currentSpend > tenant.MonthlyPlanUSD {
		deficit := currentSpend - tenant.MonthlyPlanUSD
		message := fmt.Sprintf(
			"Negative Gross Margin: Customer pays $%.2f/mo but consumed $%.2f in LLM tokens this month (deficit: -$%.2f).",
			tenant.MonthlyPlanUSD, currentSpend, deficit,
		)
		alert := store.Alert{
			TenantID:  tenant.ID,
			Type:      "negative_margin",
			Message:   message,
			Severity:  "critical",
			Timestamp: time.Now(),
		}
		_ = s.CreateAlert(alert)

		if tenant.WebhookURL != "" {
			alerts.Dispatch(tenant.WebhookURL, alerts.WebhookPayload{
				Event:     "negative_margin",
				TenantID:  tenant.ID,
				Type:      "negative_margin",
				Severity:  "critical",
				Message:   message,
				SpendUSD:  currentSpend,
				Timestamp: time.Now(),
			})
		}
		return &alert, nil
	}
	return nil, nil
}

// GenerateRecommendations inspects current usage patterns and suggests model rightsizing.
func GenerateRecommendations(s *store.Store, month string) ([]store.Recommendation, error) {
	metrics, err := s.GetTenantMetrics(month)
	if err != nil {
		return nil, err
	}

	existingRecs, _ := s.ListRecommendations(nil)
	existingKey := make(map[string]bool)
	for _, r := range existingRecs {
		key := fmt.Sprintf("%s:%s:%s", r.TenantID, r.FeatureTag, r.CurrentModel)
		existingKey[key] = true
	}

	var generated []store.Recommendation

	for _, tm := range metrics {
		// Look for tenants with negative margin or high token expenditure
		features, err := s.GetFeatureBreakdown(tm.TenantID, month)
		if err != nil {
			continue
		}

		for _, f := range features {
			model := strings.ToLower(f.TopModel)
			if model == "" {
				continue
			}

			// Match against downgrade catalogue (exact match first)
			var opt *DowngradeOption
			if val, ok := downgradeCatalog[model]; ok {
				v := val
				opt = &v
			} else {
				for key, val := range downgradeCatalog {
					if strings.HasPrefix(model, key) || strings.HasPrefix(key, model) {
						v := val
						opt = &v
						break
					}
				}
			}

			if opt == nil {
				continue
			}

			// Calculate projected savings
			savingsUSD := f.TotalCostUSD * opt.SavingsRatio
			if savingsUSD < 1.0 {
				continue // Don't bother with tiny cents
			}

			dedupKey := fmt.Sprintf("%s:%s:%s", tm.TenantID, f.FeatureTag, model)
			if existingKey[dedupKey] {
				continue
			}

			reason := opt.ReasonTemplate
			if tm.IsUnprofitable {
				reason = fmt.Sprintf("Tenant '%s' has negative gross margin (-$%.2f). %s", tm.TenantName, -tm.GrossMarginUSD, opt.ReasonTemplate)
			}

			rec := store.Recommendation{
				TenantID:          tm.TenantID,
				FeatureTag:        f.FeatureTag,
				CurrentModel:      model,
				SuggestedModel:    opt.SuggestedModel,
				MonthlySavingsUSD: savingsUSD,
				Reason:            reason,
				Applied:           false,
				Timestamp:         time.Now(),
			}

			if err := s.CreateRecommendation(rec); err == nil {
				existingKey[dedupKey] = true
				generated = append(generated, rec)
			}
		}
	}

	return generated, nil
}

// GetCatalogOption returns a recommendation option for a given model if available.
func GetCatalogOption(model string) (DowngradeOption, bool) {
	m := strings.ToLower(model)
	for k, v := range downgradeCatalog {
		if strings.HasPrefix(m, k) || strings.HasPrefix(k, m) {
			return v, true
		}
	}
	return DowngradeOption{}, false
}

// CalculateProjectedSavings calculates the exact projected savings between two models for given token counts.
func CalculateProjectedSavings(currentModel, suggestedModel string, promptTokens, completionTokens, cachedTokens int) float64 {
	currentCost := pricing.ComputeCost(currentModel, promptTokens, completionTokens, cachedTokens)
	suggestedCost := pricing.ComputeCost(suggestedModel, promptTokens, completionTokens, cachedTokens)
	savings := currentCost - suggestedCost
	if savings < 0 {
		return 0
	}
	return savings
}
