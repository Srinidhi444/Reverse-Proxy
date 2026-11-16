package health

import (
	"sync"

	"github.com/Srinidhi444/Reverse-Proxy/internal/upstream"
)

type Manager struct {
	states    map[string]*HealthState
	endpoints map[string]*upstream.Endpoint
	mu        sync.RWMutex

	failureThreshold  int
	recoveryThreshold int
}

func NewManager(failureThreshold, recoveryThreshold int) *Manager {
	if failureThreshold <= 0 {
		panic("failure threshold must be greater than zero")
	}
	if recoveryThreshold <= 0 {
		panic("recovery threshold must be greater than zero")
	}

	return &Manager{
		states:            make(map[string]*HealthState),
		endpoints:         make(map[string]*upstream.Endpoint),
		failureThreshold:  failureThreshold,
		recoveryThreshold: recoveryThreshold,
	}
}

func (m *Manager) Register(endpoint *upstream.Endpoint) {
	if endpoint == nil || endpoint.ID == "" {
		return
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.states[endpoint.ID]; exists {
		return
	}

	m.endpoints[endpoint.ID] = endpoint
	m.states[endpoint.ID] = &HealthState{
		Status:            Unhealthy,
		FailureThreshold:  m.failureThreshold,
		RecoveryThreshold: m.recoveryThreshold,
	}
}

func (m *Manager) RecordSuccess(endpointID string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	state, exists := m.states[endpointID]
	if !exists {
		return
	}

	state.RecordSuccess()
}

func (m *Manager) RecordFailure(endpointID string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	state, exists := m.states[endpointID]
	if !exists {
		return
	}

	state.RecordFailure()
}

func (m *Manager) IsHealthy(endpointID string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	state, exists := m.states[endpointID]
	if !exists {
		return false
	}

	return state.Status == Healthy
}

func (m *Manager) HealthyEndpoints() []*upstream.Endpoint {
	m.mu.RLock()
	defer m.mu.RUnlock()

	healthy := make([]*upstream.Endpoint, 0)

	for id, state := range m.states {
		if state.Status == Healthy {
			healthy = append(healthy, m.endpoints[id])
		}
	}

	return healthy
}
