package runtime

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"time"

	"github.com/Srinidhi444/Reverse-Proxy/internal/balancer"
	"github.com/Srinidhi444/Reverse-Proxy/internal/config"
	"github.com/Srinidhi444/Reverse-Proxy/internal/health"
	"github.com/Srinidhi444/Reverse-Proxy/internal/upstream"
)

const (
	failureThreshold  = 3
	recoveryThreshold = 2

	healthCheckInterval = 5 * time.Second
	healthCheckTimeout  = 2 * time.Second
	healthCheckPath     = "/health"
	healthCheckWorkers  = 10
)

type Runtime struct {
	Services      map[string]*upstream.Service
	HealthManager *health.Manager
	HealthChecker *health.Checker
}

func Build(cfg *config.Config) (*Runtime, error) {
	if cfg == nil {
		return nil, fmt.Errorf("configuration must not be nil")
	}

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	healthManager := health.NewManager(
		failureThreshold,
		recoveryThreshold,
	)

	rt := &Runtime{
		Services:      make(map[string]*upstream.Service, len(cfg.Services)),
		HealthManager: healthManager,
	}

	var allEndpoints []*upstream.Endpoint

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
			allEndpoints = append(allEndpoints, endpoint)
		}

		serviceBalancer, err := buildBalancer(
			config.BalanceType(serviceConfig.LoadBalancer),
			service.Endpoints,
			healthManager,
		)
		if err != nil {
			return nil, fmt.Errorf("service %q: %w", name, err)
		}

		service.Balancer = serviceBalancer
		rt.Services[name] = service
	}

	// NewChecker registers all endpoints with the health manager.
	rt.HealthChecker = health.NewChecker(
		healthManager,
		allEndpoints,
		healthCheckInterval,
		healthCheckTimeout,
		healthCheckPath,
		healthCheckWorkers,
	)

	return rt, nil
}

// StartHealthChecks starts the checker and waits for its initial cycle.
func (rt *Runtime) StartHealthChecks(ctx context.Context) error {
	ready := make(chan struct{})

	go rt.HealthChecker.Start(ctx, ready)

	select {
	case <-ready:
		return ctx.Err()
	case <-ctx.Done():
		return ctx.Err()
	}
}

func newTransport() *http.Transport {
	dialer := &net.Dialer{
		Timeout:   5 * time.Second,
		KeepAlive: 30 * time.Second,
	}

	transport := http.DefaultTransport.(*http.Transport).Clone()

	transport.MaxIdleConns = 100
	transport.MaxIdleConnsPerHost = 20
	transport.MaxConnsPerHost = 100
	transport.IdleConnTimeout = 90 * time.Second
	transport.ResponseHeaderTimeout = 10 * time.Second
	transport.DialContext = dialer.DialContext
	transport.TLSHandshakeTimeout = 5 * time.Second

	transport.ForceAttemptHTTP2 = false

	return transport
}

// CloseIdleConnections closes idle connections held by service transports.
func (rt *Runtime) CloseIdleConnections() {
	for _, service := range rt.Services {
		service.Transport.CloseIdleConnections()
	}
}

func buildBalancer(
	strategy config.BalanceType,
	endpoints []*upstream.Endpoint,
	healthProvider upstream.HealthProvider,
) (upstream.Balancer, error) {
	switch strategy {
	case config.RoundRobin:
		return balancer.NewRoundRobin(endpoints, healthProvider), nil

	case config.LeastConnections:
		return balancer.NewLeastConnections(endpoints, healthProvider), nil

	default:
		return nil, fmt.Errorf(
			"unsupported load balancing strategy: %q",
			strategy,
		)
	}
}
