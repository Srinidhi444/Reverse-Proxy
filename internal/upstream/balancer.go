package upstream

type Balancer interface {
	Acquire() (*Lease, error)
}
