package cache

import (
	"testing"
	"time"
)

func TestCacheOperations(t *testing.T) {
	c := NewMemoryCache(100 * time.Millisecond)

	key := HashKey("acme", "gpt-4o", []byte(`{"messages":[{"role":"user","content":"test"}]}`))
	item := Item{
		ResponseBytes:    []byte(`{"choices":[{"message":{"content":"hello"}}]}`),
		ContentType:      "application/json",
		PromptTokens:     10,
		CompletionTokens: 5,
		CachedTokens:     0,
	}

	// 1. Miss on empty cache
	if got := c.Get(key); got != nil {
		t.Errorf("expected nil for non-existent key")
	}

	// 2. Set and Hit
	c.Set(key, item, 50*time.Millisecond)
	got := c.Get(key)
	if got == nil {
		t.Fatalf("expected cache hit")
	}
	if string(got.ResponseBytes) != string(item.ResponseBytes) {
		t.Errorf("mismatched cached response: %s", string(got.ResponseBytes))
	}

	// Record savings
	c.RecordSavings(0.005)

	stats := c.GetStats()
	if stats.Hits != 1 || stats.Misses != 1 {
		t.Errorf("unexpected stats: hits=%d, misses=%d", stats.Hits, stats.Misses)
	}
	if stats.TotalSavedUSD != 0.005 {
		t.Errorf("unexpected saved USD: %f", stats.TotalSavedUSD)
	}

	// 3. Expiration check
	time.Sleep(60 * time.Millisecond)
	if expired := c.Get(key); expired != nil {
		t.Errorf("expected item to expire")
	}
}
