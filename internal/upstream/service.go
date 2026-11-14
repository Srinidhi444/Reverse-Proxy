package upstream

import "net/http"

type Service struct {
	Name      string
	Endpoints []*Endpoint
	Transport *http.Transport
	Balancer  Balancer
}
