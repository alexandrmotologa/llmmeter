# llmmeter

Multi-provider LLM cost attribution, tenant chargeback, and gross margin analyzer. Built as a single Go binary with an embedded React 19 analytics dashboard.

## Overview

Software companies that sell AI features on flat subscription plans often lose money on high-usage accounts without realizing it. A customer paying $49 per month might consume $83 worth of Claude 3.5 Sonnet tokens, creating negative gross margins.

`llmmeter` runs as a reverse proxy in front of upstream AI providers (OpenAI, Anthropic, Google Gemini, Groq, Mistral, AWS Bedrock). It extracts tenant and feature metadata from incoming HTTP headers, logs token counts, calculates costs per request, and updates customer gross margin metrics in real time.

## Key Capabilities

- **Sub-3ms Reverse Proxy**: Forwards requests to OpenAI, Anthropic, Gemini, Groq, Mistral, and Bedrock with streaming SSE passthrough.
- **Tenant Cost Attribution**: Tracks token volumes and dollars per tenant using `X-Tenant-ID` or prefixed API keys (`sk-t{tenant}-{key}`).
- **Feature Cost Attribution**: Identifies which specific features (e.g. search, summarization, copilot) drive the monthly cloud invoice via `X-Feature-Tag`.
- **Gross Margin Engine**: Compares customer subscription tiers against exact LLM token costs:
  `Gross Margin % = ((Subscription Revenue - LLM Cost) / Subscription Revenue) * 100`
- **Anomaly Detection**: Flags individual requests that cost more than 3x the tenant 30-day baseline.
- **Budget Guardrails**: Enforces monthly budget caps per customer. Throttles requests with HTTP 429 when limits are reached.
- **Downgrade Advisor**: Scans routes with negative margins and suggests cheaper model substitutions (e.g. Claude 3.5 Sonnet to Claude 3.5 Haiku) with projected monthly savings.
- **Embedded Web Cockpit**: Ships with a dark-theme React 19 dashboard inside the Go binary. Uses under 20MB of RAM.

## Architecture

```
       +--------------------------------------------------------+
       |             SaaS Backend Application                   |
       |  (Routes requests with `X-Tenant-ID: acme-corp`)       |
       +--------------------------------------------------------+
                                   |
                                   | HTTP / SSE streaming
                                   v
       +--------------------------------------------------------+
       |                    llmmeter Proxy                      |
       |             (Single Go Binary, Port 8080)              |
       +--------------------------------------------------------+
                                   |
                     +-------------+-------------+
                     |                           |
                     v                           v
     +-------------------------------+   +-------------------------------+
     |     Live Token & Cost Meter   |   |     Upstream Provider Gateway |
     | - Input / Output / Cached Tok |   | - OpenAI / Anthropic / Gemini |
     | - Live Price Matrix           |   | - SSE Chunk Passthrough       |
     | - Computes Exact USD Cost     |   | - Zero Buffering Latency      |
     +-------------------------------+   +-------------------------------+
                     |                           |
                     +-------------+-------------+
                                   |
                                   v
       +--------------------------------------------------------+
       |          SQLite FinOps Storage (WAL Mode)              |
       |  - Usage Records & Cost Aggregates                     |
       |  - Tenant Gross Margin Tables                          |
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

## Quick Start

### 1. Build and Run

```bash
# Clone the repository
git clone https://github.com/alexandrmotologa/llmmeter.git
cd llmmeter

# Build the binary (includes embedded frontend)
make build

# Seed realistic demo data
./bin/llmmeter demo --db ./finops.db

# Start proxy and analytics dashboard
./bin/llmmeter serve --port 8080 --db ./finops.db
```

Open `http://localhost:8080` in your browser to view the FinOps dashboard.

### 2. Route LLM Requests Through llmmeter

Point your OpenAI or Anthropic SDK base URL to `http://localhost:8080/v1` and pass tenant headers:

```bash
curl http://localhost:8080/v1/chat/completions \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $OPENAI_API_KEY" \
  -H "X-Tenant-ID: acme-corp" \
  -H "X-Feature-Tag: smart-summary" \
  -d '{
    "model": "gpt-4o-mini",
    "messages": [{"role": "user", "content": "Summarize monthly financials."}]
  }'
```

### 3. Generate Executive Margin Audit Report

```bash
./bin/llmmeter report --month 2026-11
```

Terminal output:

```
Executive Customer Gross Margin Audit Report — Month: 2026-11

TENANT ID     NAME                     SUBSCRIPTION   LLM COST   GROSS MARGIN   MARGIN %   STATUS
---------     ----                     ------------   --------   ------------   --------   ------
finscale-ai   FinScale Analytics       $499.00        $83.49     $415.51        83.3%      HEALTHY
cloudnova     CloudNova Systems        $199.00        $44.00     $155.00        77.9%      HEALTHY
zapmetrics    ZapMetrics Inc           $99.00         $40.82     $58.18         58.8%      HEALTHY
datapulse     DataPulse Technologies   $29.00         $39.60     $-10.60        -36.6%     NEGATIVE MARGIN (BLEED!)
acme-corp     Acme Corporation         $49.00         $38.74     $10.26         20.9%      THIN MARGIN

------------------------------------------------------------------------
Total Invoiced SaaS Revenue:  $875.00
Total LLM Infrastructure:     $246.65
Portfolio AI Gross Margin:    $628.35 (71.8%)
Negative Margin Customers:    1 / 5
------------------------------------------------------------------------
```

### 4. Export Negative Margin Accounts to CSV

```bash
./bin/llmmeter export-negative-margins --output ./unprofitable_tenants.csv
```

## CLI Reference

| Command | Description |
| --- | --- |
| `llmmeter serve [--port 8080] [--db ./finops.db]` | Starts reverse proxy and web UI |
| `llmmeter report [--month YYYY-MM] [--db ./finops.db]` | Displays executive margin audit table |
| `llmmeter export-negative-margins [--output file.csv]` | Exports unprofitable accounts to CSV |
| `llmmeter sync-pricing` | Displays current token pricing matrix for 35+ models |
| `llmmeter demo [--db ./finops.db]` | Seeds enterprise sample data for evaluation |
| `llmmeter version` | Displays version and build runtime details |

## Supported Providers & Models

- **OpenAI**: GPT-4o, GPT-4o-mini, GPT-4 Turbo, GPT-3.5 Turbo, o1, o1-mini, o3-mini, Embeddings
- **Anthropic**: Claude 3.5 Sonnet, Claude 3.5 Haiku, Claude 3 Opus
- **Google Gemini**: Gemini 1.5 Pro, Gemini 1.5 Flash, Gemini 2.0 Flash
- **Groq**: Llama 3.1 (70B, 8B), Mixtral 8x7B
- **Mistral**: Mistral Large, Mistral Small, Codestral
- **AWS Bedrock**: Claude and Llama via `bedrock/` prefix

## Storage Schema

`llmmeter` stores usage events in SQLite with Write-Ahead Logging (WAL) enabled:

```sql
CREATE TABLE tenants (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    monthly_plan_usd REAL NOT NULL DEFAULT 0.0,
    monthly_budget_usd REAL,
    webhook_url TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE request_logs (
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
```

## Documentation

- [Architecture Guide](docs/architecture.md)
- [API Reference](docs/api-reference.md)
- [Deployment & Docker Guide](docs/deployment.md)

## License

MIT License. See [LICENSE](LICENSE) for details.
