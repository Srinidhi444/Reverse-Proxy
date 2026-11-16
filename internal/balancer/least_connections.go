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
}

func NewLeastConnections(endpoints []*upstream.Endpoint) *LeastConnections {
	return &LeastConnections{
		endpoints:      append([]*upstream.Endpoint(nil), endpoints...),
		activeRequests: make(map[string]int),
		nextIndex:      0,
	}
}

func (lc *LeastConnections) Acquire() (*upstream.Endpoint, error) {
	lc.mu.Lock()
	defer lc.mu.Unlock()

	if len(lc.endpoints) == 0 {
		return nil, ErrNoEndpoints
	}

	selectedIndex := lc.nextIndex
	minimum := lc.activeRequests[lc.endpoints[selectedIndex].ID]

	for i := 1; i < len(lc.endpoints); i++ {
		index := (lc.nextIndex + i) % len(lc.endpoints)
		active := lc.activeRequests[lc.endpoints[index].ID]

		if active < minimum {
			minimum = active
			selectedIndex = index
		}
	}

	selectedEndpoint := lc.endpoints[selectedIndex]

	// Reserve the endpoint before releasing the mutex.
	lc.activeRequests[selectedEndpoint.ID]++

	// Rotate the starting position for the next tie-break.
	lc.nextIndex = (selectedIndex + 1) % len(lc.endpoints)

	return selectedEndpoint, nil
}

func (lc *LeastConnections) Release(endpoint *upstream.Endpoint) {
	if endpoint == nil {
		return
	}

	lc.mu.Lock()
	defer lc.mu.Unlock()

	if _, exists := lc.activeRequests[endpoint.ID]; !exists {
		return
	}

	if lc.activeRequests[endpoint.ID] > 0 {
		lc.activeRequests[endpoint.ID]--
	}
}
