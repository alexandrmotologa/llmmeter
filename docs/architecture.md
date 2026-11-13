# llmmeter Architecture

`llmmeter` is a high-performance reverse proxy and AI FinOps engine. It sits between SaaS backend services and upstream LLM providers (OpenAI, Anthropic, Google Gemini, Groq, Mistral, AWS Bedrock).

## Core Design Principles

1. **Sub-3ms Overhead**: Written in Go using streaming IO and SQLite in WAL mode. Token attribution happens asynchronously after the initial response headers flush.
2. **Single Self-Contained Binary**: The React 19 analytics dashboard is compiled and embedded into the Go executable via `embed.FS`. The resulting binary uses less than 20MB of RAM.
3. **Transparent Protocol Passthrough**: Upstream OpenAI and Anthropic API endpoints receive raw payloads and headers without translation overhead, preserving SSE streaming behavior.

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

## System Components

### 1. Reverse Proxy Layer (`internal/proxy`)
- Intercepts requests sent to `/v1/chat/completions` and `/v1/embeddings`.
- Extracts tenant and attribution context:
  - `X-Tenant-ID`: Identifies the customer or organization.
  - `X-Feature-Tag` or `X-Feature-ID`: Identifies the calling feature (e.g. `smart-summary`, `copilot`).
  - `X-User-ID`: End user identifier.
  - API key prefix parsing: Recognizes keys structured as `sk-t{tenantId}-{realKey}`.
- Inspects streaming SSE lines for token usage blocks or estimates usage from body length.

### 2. Pricing Engine (`internal/pricing`)
- Tracks per-million token pricing across 35+ models.
- Differentiates prompt tokens, cached prompt tokens, and completion tokens.
- Calculates exact costs according to vendor price schedules.

### 3. FinOps Engine (`internal/finops`)
- Calculates gross margin per customer:
  Gross Margin = (Subscription Revenue - LLM Cost) / Subscription Revenue * 100
- Flags accounts with negative gross margin.
- Detects consumption anomalies when a single request exceeds three times the 30-day baseline.
- Enforces monthly budget caps. When a tenant exceeds their cap, the proxy returns HTTP 429 with `X-Budget-Exceeded: true`.
- Evaluates routes and proposes model rightsizing (e.g. Claude 3.5 Sonnet to Claude 3.5 Haiku) with projected monthly savings.

### 4. Semantic Cache & Token ROI (`internal/cache`)
- In-memory SHA-256 semantic hash cache for non-streaming completions with configurable TTL.
- Serves repeated queries with sub-millisecond latency and header `X-Cache: HIT`.
- Computes dollars saved based on full prompt and completion token costs.

### 5. Rate Limiter & Circuit Breaker (`internal/ratelimit`)
- Per-tenant sliding window tracking RPM (requests per minute) and TPM (tokens per minute).
- Runaway agent loop circuit breaker trips when short-term spend surges beyond safety thresholds, returning HTTP 429 with `X-Circuit-Breaker: tripped`.

### 6. Storage Engine (`internal/store`)
- Embedded SQLite using `modernc.org/sqlite` without CGO.
- Operates in WAL mode with normal synchronization for concurrent reads and writes.
- Manages tables for tenants, request logs, alerts, and recommendations.

### 7. Observability & Prometheus Metrics (`internal/metrics`)
- Exposes native Prometheus format on `/metrics` for scraping by Prometheus, Grafana Agent, or Datadog.

### 8. Web Studio (`web/`)
- Built with React 19, TypeScript, Tailwind CSS, and Recharts.
- Provides real-time visibility into customer unit economics, feature cost waterfalls, model distribution, alerts, model rightsizing, prompt cache ROI analytics, and interactive what-if financial projections.
