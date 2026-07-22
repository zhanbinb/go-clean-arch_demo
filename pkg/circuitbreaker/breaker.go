// Package circuitbreaker wraps sony/gobreaker with a minimal, opinionated
// configuration suitable for protecting outbound calls (DB, external APIs).
package circuitbreaker

import (
	"errors"
	"time"

	"github.com/sony/gobreaker"
)

// ErrOpen is returned when the breaker is in the open state and rejects calls.
var ErrOpen = errors.New("circuit breaker is open")

// Breaker is a thin wrapper around gobreaker.CircuitBreaker.
type Breaker struct {
	cb *gobreaker.CircuitBreaker
}

// Settings configures a Breaker. Zero values fall back to sensible defaults.
type Settings struct {
	Name               string
	MaxRequests        uint32        // half-open: allowed probes
	Interval           time.Duration // closed: counters reset interval
	Timeout            time.Duration // open -> half-open duration
	ConsecutiveFails   uint32        // trip threshold
	OnStateChange      func(name string, from, to gobreaker.State)
}

// New creates a Breaker with the given settings.
func New(s Settings) *Breaker {
	if s.MaxRequests == 0 {
		s.MaxRequests = 1
	}
	if s.Interval == 0 {
		s.Interval = 60 * time.Second
	}
	if s.Timeout == 0 {
		s.Timeout = 30 * time.Second
	}
	threshold := s.ConsecutiveFails
	if threshold == 0 {
		threshold = 5
	}
	cb := gobreaker.NewCircuitBreaker(gobreaker.Settings{
		Name:        s.Name,
		MaxRequests: s.MaxRequests,
		Interval:    s.Interval,
		Timeout:     s.Timeout,
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			return counts.ConsecutiveFailures > threshold
		},
		OnStateChange: s.OnStateChange,
	})
	return &Breaker{cb: cb}
}

// Do executes fn under breaker supervision. Business errors (ErrBusiness)
// should NOT count toward the trip threshold; pass them via IsBusinessError.
func (b *Breaker) Do(fn func() (interface{}, error)) (interface{}, error) {
	v, err := b.cb.Execute(fn)
	if errors.Is(err, gobreaker.ErrOpenState) || errors.Is(err, gobreaker.ErrTooManyRequests) {
		return nil, ErrOpen
	}
	return v, err
}

// State returns the current breaker state (closed/open/half-open).
func (b *Breaker) State() gobreaker.State { return b.cb.State() }

// Counts returns the current counter snapshot (useful for /debug endpoints).
func (b *Breaker) Counts() gobreaker.Counts { return b.cb.Counts() }
