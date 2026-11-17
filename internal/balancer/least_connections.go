package balancer

import (
	"sync"

	"github.com/Srinidhi444/Reverse-Proxy/internal/upstream"
)

type LeastConnections struct {
	endpoints      []*upstream.Endpoint
	activeRequests map[string]int
	mu             sync.Mutex
	nextIndex      int
	healthProvider upstream.HealthProvider
}

func NewLeastConnections(
	endpoints []*upstream.Endpoint,
	healthProvider upstream.HealthProvider,
) *LeastConnections {
	return &LeastConnections{
		endpoints:      append([]*upstream.Endpoint(nil), endpoints...),
		activeRequests: make(map[string]int),
		healthProvider: healthProvider,
	}
}

func (lc *LeastConnections) Acquire() (*upstream.Lease, error) {
	lc.mu.Lock()

	if len(lc.endpoints) == 0 {
		lc.mu.Unlock()
		return nil, ErrNoEndpoints
	}

	var selectedEndpoint *upstream.Endpoint
	minimum := int(^uint(0) >> 1)

	// Find the healthy endpoint with the fewest active requests.
	for i := 0; i < len(lc.endpoints); i++ {
		index := (lc.nextIndex + i) % len(lc.endpoints)
		endpoint := lc.endpoints[index]

		if lc.healthProvider != nil &&
			!lc.healthProvider.IsHealthy(endpoint.ID) {
			continue
		}

		active := lc.activeRequests[endpoint.ID]

		if active < minimum {
			minimum = active
			selectedEndpoint = endpoint
		}
	}

	if selectedEndpoint == nil {
		lc.mu.Unlock()
		return nil, ErrNoEndpoints
	}

	// Reserve the endpoint before releasing the mutex.
	lc.activeRequests[selectedEndpoint.ID]++

	// Rotate the starting position for the next tie-break.
	for i, endpoint := range lc.endpoints {
		if endpoint.ID == selectedEndpoint.ID {
			lc.nextIndex = (i + 1) % len(lc.endpoints)
			break
		}
	}

	lc.mu.Unlock()

	// Release the reservation when the request finishes.
	return upstream.NewLease(selectedEndpoint, func() {
		lc.mu.Lock()
		defer lc.mu.Unlock()

		if lc.activeRequests[selectedEndpoint.ID] > 0 {
			lc.activeRequests[selectedEndpoint.ID]--
		}
	}), nil
}
