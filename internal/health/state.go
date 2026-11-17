package health

type HealthStatus string

const (
	Unknown   HealthStatus = "unknown"
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
	// The first successful check establishes initial health.
	if h.Status == Unknown {
		h.Status = Healthy
		h.ConsecutiveSuccesses = 0
		h.ConsecutiveFailures = 0
		return
	}

	h.ConsecutiveSuccesses++
	h.ConsecutiveFailures = 0

	if h.ConsecutiveSuccesses >= h.RecoveryThreshold {
		h.Status = Healthy
	}
}

func (h *HealthState) RecordFailure() {
	// The first failed check establishes initial unhealthy status.
	if h.Status == Unknown {
		h.Status = Unhealthy
		h.ConsecutiveFailures = 0
		h.ConsecutiveSuccesses = 0
		return
	}

	h.ConsecutiveFailures++
	h.ConsecutiveSuccesses = 0

	if h.ConsecutiveFailures >= h.FailureThreshold {
		h.Status = Unhealthy
	}
}
