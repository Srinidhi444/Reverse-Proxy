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
		endpoints:    slices.Clone(endpoints),
		currentIndex: 0,
	}
}

func (rr *RoundRobin) Next() (*upstream.Endpoint, error) {
	rr.mu.Lock()
	defer rr.mu.Unlock()

	if len(rr.endpoints) == 0 {
		return nil, ErrNoEndpoints
	}

	result := rr.endpoints[rr.currentIndex]
	rr.currentIndex = (rr.currentIndex + 1) % len(rr.endpoints)

	return result, nil
}
