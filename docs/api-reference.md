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
