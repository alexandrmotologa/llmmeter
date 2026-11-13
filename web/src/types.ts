export interface OverviewSummary {
  total_revenue_usd: number;
  total_cost_usd: number;
  net_margin_usd: number;
  net_margin_pct: number;
  active_tenants_count: number;
  unprofitable_count: number;
  total_tokens_month: number;
  active_alerts_count: number;
}

export interface TenantMetric {
  tenant_id: string;
  tenant_name: string;
  monthly_plan_usd: number;
  monthly_budget_usd?: number;
  total_cost_usd: number;
  gross_margin_usd: number;
  gross_margin_pct: number;
  total_tokens: number;
  request_count: number;
  is_unprofitable: boolean;
}

export interface FeatureMetric {
  feature_tag: string;
  total_cost_usd: number;
  total_tokens: number;
  request_count: number;
  avg_cost_usd: number;
  top_model: string;
}

export interface ModelMetric {
  model: string;
  provider: string;
  total_cost_usd: number;
  total_tokens: number;
  request_count: number;
}

export interface Alert {
  id: string;
  tenant_id: string;
  type: string; // anomaly, budget, negative_margin
  message: string;
  severity: string; // info, warning, critical
  resolved: boolean;
  timestamp: string;
}

export interface Recommendation {
  id: string;
  tenant_id: string;
  feature_tag: string;
  current_model: string;
  suggested_model: string;
  monthly_savings_usd: number;
  reason: string;
  applied: boolean;
  timestamp: string;
}

export interface RequestLog {
  id: string;
  tenant_id: string;
  feature_tag: string;
  user_id?: string;
  provider: string;
  model: string;
  prompt_tokens: number;
  completion_tokens: number;
  cached_tokens: number;
  cost_usd: number;
  latency_ms: number;
  timestamp: string;
}

export interface TenantDetailResponse {
  tenant: {
    id: string;
    name: string;
    monthly_plan_usd: number;
    monthly_budget_usd?: number;
    webhook_url?: string;
    created_at: string;
  };
  daily: Array<{
    date: string;
    total_cost_usd: number;
    total_tokens: number;
    request_count: number;
  }>;
  features: FeatureMetric[];
  models: ModelMetric[];
  logs: RequestLog[];
}
