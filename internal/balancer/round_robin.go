package balancer

import (
	"errors"
	"slices"
	"sync"

	"github.com/Srinidhi444/Reverse-Proxy/internal/upstream"
)

var ErrNoEndpoints = errors.New("no endpoints available")

type RoundRobin struct {
	endpoints    []*upstream.Endpoint
	mu           sync.Mutex
	currentIndex int
}

func NewRoundRobin(endpoints []*upstream.Endpoint) *RoundRobin {
	return &RoundRobin{
		endpoints: slices.Clone(endpoints),
	}
}

func (rr *RoundRobin) Acquire() (*upstream.Endpoint, error) {
	rr.mu.Lock()
	defer rr.mu.Unlock()

	if len(rr.endpoints) == 0 {
		return nil, ErrNoEndpoints
	}

	endpoint := rr.endpoints[rr.currentIndex]
	rr.currentIndex = (rr.currentIndex + 1) % len(rr.endpoints)

	return endpoint, nil
}

func (rr *RoundRobin) Release(endpoint *upstream.Endpoint) {
	// No action needed for round-robin; endpoints are always available.
}
