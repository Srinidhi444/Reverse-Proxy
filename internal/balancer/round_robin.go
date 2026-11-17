package balancer

import (
	"errors"
	"slices"
	"sync"

	"github.com/Srinidhi444/Reverse-Proxy/internal/upstream"
)

var ErrNoEndpoints = errors.New("no healthy endpoints available")

type RoundRobin struct {
	endpoints      []*upstream.Endpoint
	healthProvider upstream.HealthProvider
	mu             sync.Mutex
	currentIndex   int
}

func NewRoundRobin(
	endpoints []*upstream.Endpoint,
	healthProvider upstream.HealthProvider,
) *RoundRobin {
	return &RoundRobin{
		endpoints:      slices.Clone(endpoints),
		healthProvider: healthProvider,
	}
}

func (rr *RoundRobin) Acquire() (*upstream.Lease, error) {
	rr.mu.Lock()
	defer rr.mu.Unlock()

	if len(rr.endpoints) == 0 {
		return nil, ErrNoEndpoints
	}

	// Find the next healthy endpoint, preserving configured order.
	for i := 0; i < len(rr.endpoints); i++ {
		index := (rr.currentIndex + i) % len(rr.endpoints)
		endpoint := rr.endpoints[index]

		if rr.healthProvider == nil ||
			rr.healthProvider.IsHealthy(endpoint.ID) {

			rr.currentIndex = (index + 1) % len(rr.endpoints)
			return upstream.NewLease(endpoint, nil), nil
		}
	}

	return nil, ErrNoEndpoints
}
