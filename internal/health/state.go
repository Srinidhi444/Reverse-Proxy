package health

type HealthStatus string

const (
	Healthy   HealthStatus = "healthy"
	Unhealthy HealthStatus = "unhealthy"
)

type HealthState struct {
	Status               HealthStatus
	ConsecutiveFailures  int
	ConsecutiveSuccesses int
	FailureThreshold     int
	RecoveryThreshold    int
}

func (h *HealthState) RecordSuccess() {
	h.ConsecutiveSuccesses++
	h.ConsecutiveFailures = 0
	if h.ConsecutiveSuccesses >= h.RecoveryThreshold {
		h.Status = Healthy
	}
}

func (h *HealthState) RecordFailure() {
	h.ConsecutiveFailures++
	h.ConsecutiveSuccesses = 0
	if h.ConsecutiveFailures >= h.FailureThreshold {
		h.Status = Unhealthy
	}
}
