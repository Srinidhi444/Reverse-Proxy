package upstream

type Balancer interface {
	Next() (*Endpoint, error)
}
