package circuitbreaker

import (
	"fmt"
	"sync"
	"time"
)

type Manager struct {
	breakers map[string]*Breaker
	mu       sync.RWMutex
}

func NewManager(
	endpointIDs []string,
	failureThreshold int,
	successThreshold int,
	openDuration time.Duration,
	maxHalfOpenRequests int,
) (*Manager, error) {
	manager := &Manager{
		breakers: make(map[string]*Breaker, len(endpointIDs)),
	}

	for _, endpointID := range endpointIDs {
		if endpointID == "" {
			return nil, fmt.Errorf("endpoint ID cannot be empty")
		}

		if _, exists := manager.breakers[endpointID]; exists {
			return nil, fmt.Errorf("duplicate endpoint ID: %q", endpointID)
		}

		breaker, err := NewBreaker(
			failureThreshold,
			successThreshold,
			openDuration,
			maxHalfOpenRequests,
		)
		if err != nil {
			return nil, fmt.Errorf(
				"failed to create breaker for endpoint %q: %w",
				endpointID,
				err,
			)
		}

		manager.breakers[endpointID] = breaker
	}

	return manager, nil
}

// Allow checks whether requests can be sent to an endpoint.
func (m *Manager) Allow(endpointID string) (*Permit, bool) {
	m.mu.RLock()
	breaker, exists := m.breakers[endpointID]
	m.mu.RUnlock()

	if !exists {
		return nil, false
	}

	return breaker.Allow()
}

// State returns the current state of an endpoint's circuit breaker.
func (m *Manager) State(endpointID string) (State, bool) {
	m.mu.RLock()
	breaker, exists := m.breakers[endpointID]
	m.mu.RUnlock()

	if !exists {
		return Closed, false
	}

	return breaker.State(), true
}
