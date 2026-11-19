package circuitbreaker

import (
	"errors"
	"sync"
	"time"
)

var ErrInvalidConfiguration = errors.New("invalid circuit breaker configuration")

type Breaker struct {
	state               State
	failureCount        int
	successCount        int
	failureThreshold    int
	successThreshold    int
	openDuration        time.Duration
	openedAt            time.Time
	halfOpenInFlight    int
	maxHalfOpenRequests int

	// Invalidates results from requests admitted in an older state.
	generation uint64

	mu sync.Mutex
}

type Permit struct {
	breaker    *Breaker
	generation uint64
	state      State
	once       sync.Once
}

func NewBreaker(
	failureThreshold int,
	successThreshold int,
	openDuration time.Duration,
	maxHalfOpenRequests int,
) (*Breaker, error) {
	if failureThreshold <= 0 ||
		successThreshold <= 0 ||
		openDuration <= 0 ||
		maxHalfOpenRequests <= 0 {
		return nil, ErrInvalidConfiguration
	}

	return &Breaker{
		state:               Closed,
		failureThreshold:    failureThreshold,
		successThreshold:    successThreshold,
		openDuration:        openDuration,
		maxHalfOpenRequests: maxHalfOpenRequests,
	}, nil
}

// Allow checks whether a request can proceed and grants a permit.
func (b *Breaker) Allow() (*Permit, bool) {
	b.mu.Lock()
	defer b.mu.Unlock()

	switch b.state {
	case Closed:
		return b.newPermit(Closed), true

	case Open:
		if time.Since(b.openedAt) < b.openDuration {
			return nil, false
		}

		b.state = HalfOpen
		b.successCount = 0
		b.halfOpenInFlight = 0
		b.generation++

	case HalfOpen:
		// Continue to half-open admission check.
	}

	if b.halfOpenInFlight >= b.maxHalfOpenRequests {
		return nil, false
	}

	b.halfOpenInFlight++

	return b.newPermit(HalfOpen), true
}

func (b *Breaker) newPermit(state State) *Permit {
	return &Permit{
		breaker:    b,
		generation: b.generation,
		state:      state,
	}
}

// Complete records the result of an admitted request.
// A permit can only be completed once.
func (p *Permit) Complete(success bool) {
	if p == nil || p.breaker == nil {
		return
	}

	p.once.Do(func() {
		b := p.breaker

		b.mu.Lock()
		defer b.mu.Unlock()

		// Ignore results from requests admitted before a state transition.
		if p.generation != b.generation || p.state != b.state {
			return
		}

		switch p.state {
		case Closed:
			if success {
				b.failureCount = 0
				return
			}

			b.failureCount++

			if b.failureCount >= b.failureThreshold {
				b.open()
			}

		case HalfOpen:
			b.halfOpenInFlight--

			if !success {
				b.open()
				return
			}

			b.successCount++

			if b.successCount >= b.successThreshold {
				b.close()
			}
		}
	})
}

func (b *Breaker) open() {
	b.state = Open
	b.openedAt = time.Now()

	b.failureCount = 0
	b.successCount = 0
	b.halfOpenInFlight = 0

	b.generation++
}

func (b *Breaker) close() {
	b.state = Closed

	b.failureCount = 0
	b.successCount = 0
	b.halfOpenInFlight = 0

	b.generation++
}

// State returns the current breaker state safely.
func (b *Breaker) State() State {
	b.mu.Lock()
	defer b.mu.Unlock()

	return b.state
}
