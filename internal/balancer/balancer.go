package balancer

import "github.com/Srinidhi444/Reverse-Proxy/internal/upstream"

type Balancer interface {
	Next() (*upstream.Endpoint, error)
}
