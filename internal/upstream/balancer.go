package upstream

type Balancer interface {
	Acquire() (*Endpoint, error)
	Release(endpoint *Endpoint)
}
