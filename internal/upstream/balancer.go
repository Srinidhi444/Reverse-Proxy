package upstream

type Balancer interface {
	Acquire() (*Lease, error)
}

type HealthProvider interface {
	IsHealthy(endpointID string) bool
}
