# llmmeter API Reference

## Proxy Endpoints

### POST `/v1/chat/completions`
OpenAI-compatible chat completion proxy. Forwards requests to the target provider while metering token consumption.

#### Attribution Headers
- `X-Tenant-ID` (string, optional): Tenant identifier. Alternatively, use an API key structured as `sk-t{tenantId}-{realKey}`.
- `X-Feature-Tag` (string, optional): Name of the application feature triggering the request (e.g. `document-summary`, `search-assistant`).
- `X-User-ID` (string, optional): Specific user within the tenant.
- `X-Provider` (string, optional): Explicit provider override (`openai`, `anthropic`, `gemini`, `groq`, `mistral`, `bedrock`). If omitted, inferred from model name.
- `Authorization` (string, optional): Upstream Bearer token.

#### Budget Enforcing
If the tenant has exceeded their configured `monthly_budget_usd`, the proxy returns:
```http
HTTP/1.1 429 Too Many Requests
Content-Type: application/json
X-Budget-Exceeded: true

{
  "error": {
    "message": "Monthly LLM budget cap exceeded for this tenant.",
    "type": "budget_exceeded"
  }
}
```

---

### POST `/v1/embeddings`
OpenAI-compatible embeddings proxy. Attributes prompt tokens and costs to the calling tenant.

---

### GET `/v1/models`
Returns available models with input, output, and cached token pricing per 1,000,000 tokens.

---

### GET `/health`
Health check endpoint returning engine status and system timestamp.

```json
{
  "status": "ok",
  "engine": "llmmeter-go",
  "timestamp": "2026-11-13T20:50:00Z"
}
```

---

## FinOps REST Endpoints

### GET `/api/overview`
Returns global portfolio KPIs for the current month.
- Query Parameters: `month` (optional, format `YYYY-MM`)
- Response: `total_revenue_usd`, `total_cost_usd`, `net_margin_usd`, `net_margin_pct`, `active_tenants_count`, `unprofitable_count`, `total_tokens_month`, `active_alerts_count`.

---

### GET `/api/tenants`
Lists customer accounts with monthly revenue, LLM cost, gross margin dollars, gross margin percentage, token volume, and request counts.
- Query Parameters:
  - `month` (optional, format `YYYY-MM`)
  - `unprofitable` (optional, boolean): Set to `true` to return only accounts with negative gross margin.

---

### POST `/api/tenants`
Registers or updates a customer account.
```json
{
  "id": "acme-corp",
  "name": "Acme Corporation",
  "monthly_plan_usd": 49.00,
  "monthly_budget_usd": 75.00,
  "webhook_url": "https://hooks.slack.com/services/..."
}
```

---

### GET `/api/tenants/{id}`
Returns customer audit details including daily spending trends, feature breakdowns, and recent proxy logs.

---

### GET `/api/tenants/{id}/statement`
Generates an official billing and AI consumption audit statement.
- Query Parameters:
  - `month` (optional, format `YYYY-MM`)
  - `format` (optional, `json` or `html`): When set to `html`, returns a printable statement styled for PDF generation.
- Response includes subscription plan fee, total AI usage cost, net gross margin, token metrics, and per-feature cost distribution.

---

### GET `/api/cache-analytics`
Returns aggregate and per-feature prompt caching metrics.
- Response includes `total_prompt_tokens`, `total_cached_tokens`, `overall_hit_rate_pct`, `total_saved_usd`, and `top_features` breakdown.

---

### POST `/api/simulator/simulate`
Runs what-if financial projections based on potential pricing adjustments or model migrations.
- Request Body:
  - `subscription_increase_usd` (number): Proposed flat plan price increase.
  - `model_replacements` (object): Map of source model to replacement target model.
- Response returns baseline revenue/cost/margin versus simulated revenue/cost/margin and projected net monthly gain.

---

### GET `/metrics`
Prometheus metrics exposition endpoint. Exposes standard metrics including:
- `llmmeter_http_requests_total`
- `llmmeter_tokens_total`
- `llmmeter_cost_usd_total`
- `llmmeter_cache_hits_total`
- `llmmeter_cache_saved_usd_total`
- `llmmeter_unprofitable_tenants`

---

### DELETE `/api/tenants/{id}`
Removes a tenant record.

---

### GET `/api/features`
Returns feature spending attribution ranked by total cost.
- Query Parameters: `tenant_id` (optional), `month` (optional).

---

### GET `/api/alerts`
Returns active anomaly and budget alerts.
- Query Parameters: `resolved` (optional, boolean).

---

### POST `/api/alerts/{id}/resolve`
Marks an alert as resolved.

---

### GET `/api/recommendations`
Returns model downgrade proposals with projected monthly savings.
- Query Parameters: `applied` (optional, boolean).

---

### POST `/api/recommendations`
Triggers an automated scan of all tenant routes and generates rightsizing suggestions.

---

### POST `/api/recommendations/{id}/apply`
Marks a recommendation as applied.

---

### GET `/api/export/csv`
Generates and downloads a CSV export of customer gross margins.
