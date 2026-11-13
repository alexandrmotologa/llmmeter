package cache

import (
	"crypto/sha256"
	"encoding/hex"
	"sync"
	"sync/atomic"
	"time"
)

// Item holds a cached LLM response and token attribution metadata.
type Item struct {
	ResponseBytes    []byte
	ContentType      string
	PromptTokens     int
	CompletionTokens int
	CachedTokens     int
	ExpiresAt        time.Time
}

// Stats holds real-time cache performance counters.
type Stats struct {
	Hits           uint64  `json:"hits"`
	Misses         uint64  `json:"misses"`
	HitRatePct     float64 `json:"hit_rate_pct"`
	ItemsCount     int     `json:"items_count"`
	TotalSavedUSD  float64 `json:"total_saved_usd"`
}

// MemoryCache provides in-memory thread-safe caching with TTL.
type MemoryCache struct {
	mu            sync.RWMutex
	items         map[string]Item
	hits          uint64
	misses        uint64
	totalSavedUSD float64
	defaultTTL    time.Duration
}

// NewMemoryCache initializes an in-memory response cache.
func NewMemoryCache(defaultTTL time.Duration) *MemoryCache {
	if defaultTTL <= 0 {
		defaultTTL = 5 * time.Minute
	}
	c := &MemoryCache{
		items:      make(map[string]Item),
		defaultTTL: defaultTTL,
	}

	// Background eviction worker
	go func() {
		ticker := time.NewTicker(1 * time.Minute)
		for range ticker.C {
			c.evictExpired()
		}
	}()

	return c
}

// HashKey computes a SHA-256 hash from tenant, model, and request payload.
func HashKey(tenantID, model string, body []byte) string {
	h := sha256.New()
	h.Write([]byte(tenantID))
	h.Write([]byte("|"))
	h.Write([]byte(model))
	h.Write([]byte("|"))
	h.Write(body)
	return hex.EncodeToString(h.Sum(nil))
}

// Get looks up a cached response. Returns nil if missing or expired.
func (c *MemoryCache) Get(key string) *Item {
	c.mu.RLock()
	item, found := c.items[key]
	c.mu.RUnlock()

	if !found || time.Now().After(item.ExpiresAt) {
		atomic.AddUint64(&c.misses, 1)
		return nil
	}

	atomic.AddUint64(&c.hits, 1)
	return &item
}

// Set stores a response with a given TTL.
func (c *MemoryCache) Set(key string, item Item, ttl time.Duration) {
	if ttl <= 0 {
		ttl = c.defaultTTL
	}
	item.ExpiresAt = time.Now().Add(ttl)

	c.mu.Lock()
	c.items[key] = item
	c.mu.Unlock()
}

// RecordSavings records avoided upstream token cost.
func (c *MemoryCache) RecordSavings(savedUSD float64) {
	c.mu.Lock()
	c.totalSavedUSD += savedUSD
	c.mu.Unlock()
}

func (c *MemoryCache) evictExpired() {
	now := time.Now()
	c.mu.Lock()
	defer c.mu.Unlock()

	for k, v := range c.items {
		if now.After(v.ExpiresAt) {
			delete(c.items, k)
		}
	}
}

// GetStats returns cache hit/miss statistics.
func (c *MemoryCache) GetStats() Stats {
	c.mu.RLock()
	defer c.mu.RUnlock()

	h := atomic.LoadUint64(&c.hits)
	m := atomic.LoadUint64(&c.misses)
	total := h + m
	rate := 0.0
	if total > 0 {
		rate = (float64(h) / float64(total)) * 100.0
	}

	return Stats{
		Hits:          h,
		Misses:        m,
		HitRatePct:    rate,
		ItemsCount:    len(c.items),
		TotalSavedUSD: c.totalSavedUSD,
	}
}
