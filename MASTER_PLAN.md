# `llmmeter` — Architecture & Engineering Master Plan
> **Multi-Provider LLM Cost Attribution, Tenant Chargeback & Margin Analyzer**  
> *Target: High-Performance Single-Binary Go Proxy (< 20MB RAM) + Embedded React 19 FinOps Studio*

---

## 1. Executive Summary & Market Problem

### 1.1 The Silent Margin Bleed in AI-Powered SaaS
SaaS companies integrating generative AI features routinely price their software with flat subscription tiers (e.g., \$29/month or \$99/month). However, because AI workloads consume tokens on variable usage pricing from OpenAI, Anthropic, and Google:
1. **Negative Gross Margin Customers**: A small cohort of power users or automated API scripts run thousands of queries, consuming **\$150 to \$300/month in LLM tokens** while paying only a \$29 subscription, silently bankrupting unit economics.
2. **Zero Route & Feature Attribution**: Engineering and finance teams cannot see which specific features (e.g., *Smart Search* vs *Document Summarizer* vs *Chatbot*) drive the monthly \$25,000 cloud AI invoice.
3. **No Dynamic Model Rightsizing**: Teams continue routing simple tasks to \$15/M token frontier models when a \$0.15/M model would produce identical results.

### 1.2 The `llmmeter` Solution
`llmmeter` is a transparent, high-speed reverse proxy and FinOps engine written in Go with an embedded React 19 analytics cockpit:
- **Transparent Reverse Proxy (< 3ms Overhead)**: Sits between client applications and LLM providers (OpenAI, Anthropic, Gemini, Mistral, Bedrock).
- **Multi-Tenant Cost Attribution**: Extracts `X-Tenant-ID`, `X-Feature-Tag`, and `X-User-ID` headers, recording exact token counts and costs per tenant.
- **Unit Economics & Margin Engine**: Connects subscription tiers or billing data to calculate gross margin per customer:
  $$\text{Gross Margin} = \frac{\text{Subscription Revenue} - \text{LLM Cost}}{\text{Subscription Revenue}} \times 100\%$$
- **Anomaly Detection & Budget Safeguards**: Automatically alerts or throttles tenants when consumption exceeds predefined budget caps.
- **Model Downgrade Recommendation Engine**: Analyzes task complexity and suggests automated model cascades (e.g. switching a route from Claude 3.5 Sonnet to Haiku 3.5, saving 80%+).

---

## 2. High-Level System Architecture

```
       +--------------------------------------------------------+
       |             SaaS Backend Application                   |
       |  (Routes OpenAI / Anthropic requests through proxy)    |
       +--------------------------------------------------------+
                                   |
                                   | HTTP / SSE with `X-Tenant-ID: org_123`
                                   v
       +--------------------------------------------------------+
       |                    llmmeter Proxy                      |
       |          (Single Go Binary, Sub-3ms Overhead)          |
       +--------------------------------------------------------+
                                   |
                     +-------------+-------------+
                     |                           |
                     v                           v
     +-------------------------------+   +-------------------------------+
     |     Live Token & Cost Meter   |   |     Upstream Provider Gateway |
     | - Live Model Price Matrix     |   | - OpenAI / Anthropic / Gemini |
     | - Input / Output / Cached Tok |   | - Streaming SSE Passthrough   |
     | - Computes Exact USD Cost     |   | - Zero Buffering Latency      |
     +-------------------------------+   +-------------------------------+
                     |                           |
                     +-------------+-------------+
                                   |
                                   v
       +--------------------------------------------------------+
       |          SQLite FinOps Storage (WAL Mode)              |
       |  - Usage Records & Cost Aggregates                     |
       |  - Tenant Margin Tables                                |
       +--------------------------------------------------------+
                                   |
                                   v
       +--------------------------------------------------------+
       |            React 19 FinOps Studio & Alerts             |
       |  - Tenant Margin Heatmap (Negative Margin Alerts)      |
       |  - Feature Breakdown Waterfall                         |
       |  - 1-Click Downgrade Recommendation Optimizer          |
       +--------------------------------------------------------+
```

---

## 3. SQLite Storage Schema (`store/schema.sql`)

```sql
CREATE TABLE IF NOT EXISTS tenants (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    monthly_plan_usd REAL NOT NULL DEFAULT 0.0,
    monthly_budget_usd REAL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS request_logs (
    id TEXT PRIMARY KEY,
    tenant_id TEXT NOT NULL REFERENCES tenants(id),
    feature_tag TEXT NOT NULL,
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
```

---

## 4. CLI Command Surface

```bash
# Start the llmmeter reverse proxy and web FinOps studio
llmmeter serve --port 8080 --db ./finops.db

# Ingest or update pricing matrix for all models
llmmeter sync-pricing

# Output an executive margin audit report for current month
llmmeter report --month 2026-11 --format table

# Export negative margin tenants to CSV
llmmeter export-negative-margins --output ./unprofitable_tenants.csv
```

---

## 5. Implementation Roadmap (5 Phases)

### Phase 1: High-Speed Reverse Proxy & Pricing Engine
- Build streaming reverse proxy supporting OpenAI and Anthropic API protocols.
- Implement live token pricing engine covering all major providers and cached tokens.

### Phase 2: Tenant Attribution & SQLite Storage
- Extract tenant and feature headers; record usage events to SQLite in WAL mode.
- Build high-speed aggregate rollups (hourly, daily, monthly per tenant).

### Phase 3: Margin & Unit Economics Calculator
- Implement customer gross margin formulas and identify unprofitable accounts.
- Implement budget cap alerts via Slack and Webhooks.

### Phase 4: Model Downgrade & Optimization Advisor
- Analyze token patterns and recommend cost-saving model substitutions.
- Compute potential monthly cloud savings projections.

### Phase 5: Embedded React 19 Studio & Docs
- Build modern FinOps dashboard with margin heatmaps and feature cost waterfalls.
- Package as a portable single binary in Go under 20MB RAM.
