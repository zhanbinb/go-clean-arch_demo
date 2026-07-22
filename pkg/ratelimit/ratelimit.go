// Package ratelimit provides an in-memory token-bucket limiter, suitable for
// single-instance deployments. For multi-instance setups, swap the storage
// for Redis.
package ratelimit

import (
	"sync"
	"time"

	"golang.org/x/time/rate"
)

// Limiter holds per-key token buckets and applies cleanup.
type Limiter struct {
	mu       sync.Mutex
	visitors map[string]*visitor
	rps      rate.Limit
	burst    int
	ttl      time.Duration
}

type visitor struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

// New builds a Limiter. rps is tokens added per second; burst is the bucket size.
// cleanupInterval is how often stale visitors are evicted.
func New(rps float64, burst int, cleanupInterval time.Duration) *Limiter {
	l := &Limiter{
		visitors: make(map[string]*visitor),
		rps:      rate.Limit(rps),
		burst:    burst,
		ttl:      cleanupInterval,
	}
	go l.cleanupLoop()
	return l
}

// Allow consumes one token for the given key. Returns false if the bucket
// is empty (caller should respond with 429).
func (l *Limiter) Allow(key string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	v, ok := l.visitors[key]
	if !ok {
		v = &visitor{limiter: rate.NewLimiter(l.rps, l.burst)}
		l.visitors[key] = v
	}
	v.lastSeen = time.Now()
	return v.limiter.Allow()
}

func (l *Limiter) cleanupLoop() {
	t := time.NewTicker(l.ttl)
	defer t.Stop()
	for range t.C {
		l.evict()
	}
}

func (l *Limiter) evict() {
	l.mu.Lock()
	defer l.mu.Unlock()
	cutoff := time.Now().Add(-l.ttl)
	for k, v := range l.visitors {
		if v.lastSeen.Before(cutoff) {
			delete(l.visitors, k)
		}
	}
}
