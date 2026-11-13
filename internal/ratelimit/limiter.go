package ratelimit

import (
	"fmt"
	"sync"
	"time"
)

// TenantConfig defines rate limits and spend velocity thresholds for a tenant.
type TenantConfig struct {
	MaxRPM             int     // Max requests per minute
	MaxTPM             int     // Max tokens per minute
	TripSpendThreshold float64 // Dollars spent within 10 minutes before tripping circuit breaker
}

// Default limits
var defaultConfig = TenantConfig{
	MaxRPM:             120,    // 120 requests/minute
	MaxTPM:             300000, // 300k tokens/minute
	TripSpendThreshold: 25.0,   // $25 within 10 minutes trips circuit breaker
}

type spendEvent struct {
	costUSD   float64
	timestamp time.Time
}

type tenantState struct {
	mu           sync.Mutex
	requestTimes []time.Time
	tokenCounts  []struct {
		tokens    int
		timestamp time.Time
	}
	spendHistory       []spendEvent
	circuitBreakerTrip time.Time // If non-zero, circuit is open until this time
}

// Limiter manages rate limits and circuit breaking across all tenants.
type Limiter struct {
	mu      sync.RWMutex
	configs map[string]TenantConfig
	states  map[string]*tenantState
}

// NewLimiter creates a new rate limiter & circuit breaker.
func NewLimiter() *Limiter {
	return &Limiter{
		configs: make(map[string]TenantConfig),
		states:  make(map[string]*tenantState),
	}
}

func (l *Limiter) getOrCreateState(tenantID string) (*tenantState, TenantConfig) {
	l.mu.RLock()
	cfg, hasCfg := l.configs[tenantID]
	s, hasState := l.states[tenantID]
	l.mu.RUnlock()

	if !hasCfg {
		cfg = defaultConfig
	}

	if !hasState {
		l.mu.Lock()
		s = &tenantState{}
		l.states[tenantID] = s
		l.mu.Unlock()
	}

	return s, cfg
}

// SetTenantLimits configures custom limits for a specific tenant.
func (l *Limiter) SetTenantLimits(tenantID string, cfg TenantConfig) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.configs[tenantID] = cfg
}

// Allow checks RPM and verifies that the circuit breaker is not tripped.
func (l *Limiter) Allow(tenantID string) (bool, string) {
	state, cfg := l.getOrCreateState(tenantID)

	state.mu.Lock()
	defer state.mu.Unlock()

	now := time.Now()

	// 1. Check Circuit Breaker
	if !state.circuitBreakerTrip.IsZero() {
		if now.Before(state.circuitBreakerTrip) {
			remainingSec := int(state.circuitBreakerTrip.Sub(now).Seconds())
			return false, fmt.Sprintf("Circuit breaker active: Runaway spend detected. Throttled for %ds.", remainingSec)
		}
		// Reset trip after cooldown
		state.circuitBreakerTrip = time.Time{}
	}

	// 2. Check RPM
	cutoff := now.Add(-1 * time.Minute)
	valid := state.requestTimes[:0]
	for _, t := range state.requestTimes {
		if t.After(cutoff) {
			valid = append(valid, t)
		}
	}
	state.requestTimes = valid

	if len(state.requestTimes) >= cfg.MaxRPM {
		return false, fmt.Sprintf("Rate limit exceeded: %d/%d RPM limit reached", len(state.requestTimes), cfg.MaxRPM)
	}

	state.requestTimes = append(state.requestTimes, now)
	return true, ""
}

// RecordSpend tracks request expenditure and trips the circuit breaker if runaway spend is detected.
func (l *Limiter) RecordSpend(tenantID string, costUSD float64, tokens int) (bool, string) {
	state, cfg := l.getOrCreateState(tenantID)

	state.mu.Lock()
	defer state.mu.Unlock()

	now := time.Now()

	// Record tokens
	state.tokenCounts = append(state.tokenCounts, struct {
		tokens    int
		timestamp time.Time
	}{tokens: tokens, timestamp: now})

	// Record spend
	state.spendHistory = append(state.spendHistory, spendEvent{costUSD: costUSD, timestamp: now})

	// Check 10-minute spend window
	tenMinCutoff := now.Add(-10 * time.Minute)
	validSpend := state.spendHistory[:0]
	total10mSpend := 0.0
	for _, ev := range state.spendHistory {
		if ev.timestamp.After(tenMinCutoff) {
			validSpend = append(validSpend, ev)
			total10mSpend += ev.costUSD
		}
	}
	state.spendHistory = validSpend

	if cfg.TripSpendThreshold > 0 && total10mSpend >= cfg.TripSpendThreshold {
		// Trip circuit breaker for 15 minutes!
		state.circuitBreakerTrip = now.Add(15 * time.Minute)
		return true, fmt.Sprintf("Circuit breaker TRIPPED: Tenant spent $%.2f in <10m (threshold: $%.2f). Cooldown active.", total10mSpend, cfg.TripSpendThreshold)
	}

	return false, ""
}
