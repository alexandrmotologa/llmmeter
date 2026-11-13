package metrics

import (
	"fmt"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/alexandrmotologa/llmmeter/internal/store"
)

// Collector holds in-memory Prometheus counters.
type Collector struct {
	mu                  sync.RWMutex
	requestsTotal       map[string]uint64 // key: tenant,model,provider,status
	tokensTotal         map[string]uint64 // key: tenant,type
	costTotal           map[string]float64 // key: tenant,model
	cacheHits           uint64
	cacheMisses         uint64
	circuitBreakerTrips uint64
	store               *store.Store
}

var globalCollector *Collector
var once sync.Once

// Init initializes the global Prometheus metrics collector.
func Init(s *store.Store) *Collector {
	once.Do(func() {
		globalCollector = &Collector{
			requestsTotal: make(map[string]uint64),
			tokensTotal:   make(map[string]uint64),
			costTotal:     make(map[string]float64),
			store:         s,
		}
	})
	return globalCollector
}

// Get returns the global metrics collector.
func Get() *Collector {
	return globalCollector
}

// RecordRequest tracks a completed proxy request.
func (c *Collector) RecordRequest(tenantID, provider, model string, statusCode int, promptTokens, completionTokens, cachedTokens int, costUSD float64) {
	if c == nil {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()

	reqKey := fmt.Sprintf(`tenant="%s",provider="%s",model="%s",status="%d"`, sanitize(tenantID), sanitize(provider), sanitize(model), statusCode)
	c.requestsTotal[reqKey]++

	pKey := fmt.Sprintf(`tenant="%s",type="prompt"`, sanitize(tenantID))
	c.tokensTotal[pKey] += uint64(promptTokens)

	cKey := fmt.Sprintf(`tenant="%s",type="completion"`, sanitize(tenantID))
	c.tokensTotal[cKey] += uint64(completionTokens)

	if cachedTokens > 0 {
		caKey := fmt.Sprintf(`tenant="%s",type="cached"`, sanitize(tenantID))
		c.tokensTotal[caKey] += uint64(cachedTokens)
	}

	costKey := fmt.Sprintf(`tenant="%s",model="%s"`, sanitize(tenantID), sanitize(model))
	c.costTotal[costKey] += costUSD
}

// RecordCacheHit records a cache hit.
func (c *Collector) RecordCacheHit() {
	if c != nil {
		atomic.AddUint64(&c.cacheHits, 1)
	}
}

// RecordCacheMiss records a cache miss.
func (c *Collector) RecordCacheMiss() {
	if c != nil {
		atomic.AddUint64(&c.cacheMisses, 1)
	}
}

// RecordCircuitBreakerTrip records a circuit breaker trip.
func (c *Collector) RecordCircuitBreakerTrip() {
	if c != nil {
		atomic.AddUint64(&c.circuitBreakerTrips, 1)
	}
}

// Handler returns an http.HandlerFunc that renders Prometheus format text.
func (c *Collector) Handler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; version=0.0.4; charset=utf-8")

		var b strings.Builder

		b.WriteString("# HELP llmmeter_http_requests_total Total number of LLM proxy requests\n")
		b.WriteString("# TYPE llmmeter_http_requests_total counter\n")
		c.mu.RLock()
		for labels, count := range c.requestsTotal {
			b.WriteString(fmt.Sprintf("llmmeter_http_requests_total{%s} %d\n", labels, count))
		}

		b.WriteString("\n# HELP llmmeter_tokens_total Total tokens processed by type\n")
		b.WriteString("# TYPE llmmeter_tokens_total counter\n")
		for labels, count := range c.tokensTotal {
			b.WriteString(fmt.Sprintf("llmmeter_tokens_total{%s} %d\n", labels, count))
		}

		b.WriteString("\n# HELP llmmeter_cost_usd_total Total cost in USD per tenant and model\n")
		b.WriteString("# TYPE llmmeter_cost_usd_total counter\n")
		for labels, cost := range c.costTotal {
			b.WriteString(fmt.Sprintf("llmmeter_cost_usd_total{%s} %.6f\n", labels, cost))
		}
		c.mu.RUnlock()

		b.WriteString(fmt.Sprintf("\n# HELP llmmeter_cache_hits_total Total semantic cache hits\n# TYPE llmmeter_cache_hits_total counter\nllmmeter_cache_hits_total %d\n", atomic.LoadUint64(&c.cacheHits)))
		b.WriteString(fmt.Sprintf("\n# HELP llmmeter_cache_misses_total Total semantic cache misses\n# TYPE llmmeter_cache_misses_total counter\nllmmeter_cache_misses_total %d\n", atomic.LoadUint64(&c.cacheMisses)))
		b.WriteString(fmt.Sprintf("\n# HELP llmmeter_circuit_breaker_trips_total Total times circuit breaker tripped\n# TYPE llmmeter_circuit_breaker_trips_total counter\nllmmeter_circuit_breaker_trips_total %d\n", atomic.LoadUint64(&c.circuitBreakerTrips)))

		// Add live store gauges if available
		if c.store != nil {
			if summary, err := c.store.GetOverviewSummary(""); err == nil {
				b.WriteString(fmt.Sprintf("\n# HELP llmmeter_active_tenants Active tenant count\n# TYPE llmmeter_active_tenants gauge\nllmmeter_active_tenants %d\n", summary.ActiveTenantsCount))
				b.WriteString(fmt.Sprintf("\n# HELP llmmeter_unprofitable_tenants Unprofitable tenants count\n# TYPE llmmeter_unprofitable_tenants gauge\nllmmeter_unprofitable_tenants %d\n", summary.UnprofitableCount))
				b.WriteString(fmt.Sprintf("\n# HELP llmmeter_net_margin_pct Portfolio gross margin percentage\n# TYPE llmmeter_net_margin_pct gauge\nllmmeter_net_margin_pct %.2f\n", summary.NetMarginPct))
			}
		}

		_, _ = w.Write([]byte(b.String()))
	}
}

func sanitize(s string) string {
	s = strings.ReplaceAll(s, `"`, `\"`)
	s = strings.ReplaceAll(s, "\n", "")
	return s
}
