package upstream

import "net/http"

// Service represents an upstream service and its backend instances.
type Service struct {
	Name      string
	Endpoints []*Endpoint
	Transport *http.Transport
}
