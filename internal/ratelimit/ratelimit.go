package ratelimit

import (
	"sync"
	"time"

	"golang.org/x/time/rate"
)

type Limiter struct {
	tpsLimiter   *rate.Limiter
	tokenLimiter *rate.Limiter
}

type Manager struct {
	limiters map[string]*Limiter
	mu       sync.Mutex
}

func NewManager() *Manager {
	return &Manager{
		limiters: make(map[string]*Limiter),
	}
}

func (m *Manager) GetLimiter(key string, tps float64, tokenLimit int64) *Limiter {
	m.mu.Lock()
	defer m.mu.Unlock()

	if l, ok := m.limiters[key]; ok {
		return l
	}

	var tpsLim *rate.Limiter
	if tps > 0 {
		tpsLim = rate.NewLimiter(rate.Limit(tps), int(tps)+1)
	}

	var tokenLim *rate.Limiter
	if tokenLimit > 0 {
		limit := rate.Limit(float64(tokenLimit) / 60.0)
		tokenLim = rate.NewLimiter(limit, int(tokenLimit))
	}

	l := &Limiter{
		tpsLimiter:   tpsLim,
		tokenLimiter: tokenLim,
	}
	m.limiters[key] = l
	return l
}

func (l *Limiter) AllowTPS() bool {
	if l.tpsLimiter == nil {
		return true
	}
	return l.tpsLimiter.Allow()
}

func (l *Limiter) AllowTokens(n int) bool {
	if l.tokenLimiter == nil {
		return true
	}
	return l.tokenLimiter.AllowN(time.Now(), n)
}
