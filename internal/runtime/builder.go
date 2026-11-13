package runtime

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/Srinidhi444/Reverse-Proxy/internal/config"
	"github.com/Srinidhi444/Reverse-Proxy/internal/upstream"
)

type Runtime struct {
	Services map[string]*upstream.Service
}

func Build(cfg *config.Config) (*Runtime, error) {
	if cfg == nil {
		return nil, fmt.Errorf("configuration must not be nil")
	}

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	rt := &Runtime{
		Services: make(map[string]*upstream.Service, len(cfg.Services)),
	}

	for name, serviceConfig := range cfg.Services {
		service := &upstream.Service{
			Name:      name,
			Transport: newTransport(),
		}

		for _, endpointConfig := range serviceConfig.Endpoints {
			parsedURL, err := url.Parse(endpointConfig.URL)
			if err != nil {
				return nil, fmt.Errorf(
					"failed to parse endpoint URL %q for service %q: %w",
					endpointConfig.URL, name, err,
				)
			}

			endpoint := &upstream.Endpoint{
				ID:  name + ":" + parsedURL.String(),
				URL: parsedURL,
			}

			service.Endpoints = append(service.Endpoints, endpoint)
		}

		rt.Services[name] = service
	}

	return rt, nil
}

func newTransport() *http.Transport {
	transport := http.DefaultTransport.(*http.Transport).Clone()

	transport.ForceAttemptHTTP2 = false

	return transport
}

// CloseIdleConnections closes idle connections held by service transports.
func (rt *Runtime) CloseIdleConnections() {
	for _, service := range rt.Services {
		service.Transport.CloseIdleConnections()
	}
}
