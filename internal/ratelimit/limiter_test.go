package ratelimit

import (
	"testing"
)

func TestRateLimiterAndCircuitBreaker(t *testing.T) {
	lim := NewLimiter()
	lim.SetTenantLimits("tenant-test", TenantConfig{
		MaxRPM:             3,
		MaxTPM:             10000,
		TripSpendThreshold: 5.0, // $5 in 10 minutes trips
	})

	// 1. Allow first 3 requests
	for i := 0; i < 3; i++ {
		allowed, msg := lim.Allow("tenant-test")
		if !allowed {
			t.Fatalf("request %d should be allowed, got: %s", i+1, msg)
		}
	}

	// 2. 4th request should be blocked by RPM
	allowed, msg := lim.Allow("tenant-test")
	if allowed {
		t.Errorf("4th request should have been blocked by RPM")
	}
	if msg == "" {
		t.Errorf("expected rejection reason")
	}

	// 3. Test Circuit Breaker
	lim2 := NewLimiter()
	lim2.SetTenantLimits("tenant-runaway", TenantConfig{
		MaxRPM:             100,
		TripSpendThreshold: 10.0,
	})

	// Simulate runaway spend ($6 + $5 = $11 in <10 min)
	tripped, _ := lim2.RecordSpend("tenant-runaway", 6.0, 1000)
	if tripped {
		t.Errorf("should not trip yet at $6")
	}

	tripped, reason := lim2.RecordSpend("tenant-runaway", 5.0, 1000)
	if !tripped {
		t.Errorf("circuit breaker should have tripped at $11 spend")
	}
	if reason == "" {
		t.Errorf("expected circuit breaker reason")
	}

	// Now subsequent requests must be blocked by circuit breaker
	allowAfterTrip, tripMsg := lim2.Allow("tenant-runaway")
	if allowAfterTrip {
		t.Errorf("request should be blocked while circuit breaker is active")
	}
	if tripMsg == "" {
		t.Errorf("expected circuit breaker rejection message")
	}
}
