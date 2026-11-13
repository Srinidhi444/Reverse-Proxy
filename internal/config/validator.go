package config

import (
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"
)

type BalanceType string

const (
	RoundRobin       BalanceType = "round_robin"
	LeastConnections BalanceType = "least_connections"
)

func (cfg *Config) Validate() error {
	var validationErrors []error

	// Validate server configuration.
	if strings.TrimSpace(cfg.Server.Address) == "" {
		validationErrors = append(validationErrors,
			fmt.Errorf("server.address: must not be empty"))
	}

	validateTimeout := func(name, value string) {
		duration, err := time.ParseDuration(value)
		if err != nil {
			validationErrors = append(validationErrors,
				fmt.Errorf("server.%s: invalid duration %q: %w", name, value, err))
			return
		}

		if duration <= 0 {
			validationErrors = append(validationErrors,
				fmt.Errorf("server.%s: must be greater than zero", name))
		}
	}

	validateTimeout("read_timeout", cfg.Server.ReadTimeout)
	validateTimeout("write_timeout", cfg.Server.WriteTimeout)
	validateTimeout("idle_timeout", cfg.Server.IdleTimeout)

	// Validate services.
	if len(cfg.Services) == 0 {
		validationErrors = append(validationErrors,
			fmt.Errorf("services: at least one service is required"))
	}

	for name, service := range cfg.Services {
		if strings.TrimSpace(name) == "" {
			validationErrors = append(validationErrors,
				fmt.Errorf("services: service name must not be empty"))
		}

		switch BalanceType(service.LoadBalancer) {
		case RoundRobin, LeastConnections:
			// Supported load balancer.
		default:
			validationErrors = append(validationErrors,
				fmt.Errorf("services.%s.load_balancer: unsupported value %q",
					name, service.LoadBalancer))
		}

		if len(service.Endpoints) == 0 {
			validationErrors = append(validationErrors,
				fmt.Errorf("services.%s.endpoints: at least one endpoint is required",
					name))
		}

		for i, endpoint := range service.Endpoints {
			parsedURL, err := url.Parse(endpoint.URL)

			if err != nil ||
				parsedURL == nil ||
				(parsedURL.Scheme != "http" && parsedURL.Scheme != "https") ||
				parsedURL.Host == "" ||
				parsedURL.User != nil ||
				parsedURL.Fragment != "" {
				validationErrors = append(validationErrors,
					fmt.Errorf("services.%s.endpoints[%d].url: invalid URL %q",
						name, i, endpoint.URL))
			}
		}
	}

	// Validate routes.
	if len(cfg.Routes) == 0 {
		validationErrors = append(validationErrors,
			fmt.Errorf("routes: at least one route is required"))
	}

	seenRoutes := make(map[string]struct{})

	for i, route := range cfg.Routes {
		if route.Path == "" {
			validationErrors = append(validationErrors,
				fmt.Errorf("routes[%d].path: must not be empty", i))
		} else {
			if !strings.HasPrefix(route.Path, "/") {
				validationErrors = append(validationErrors,
					fmt.Errorf("routes[%d].path: must start with /", i))
			}

			if strings.ContainsAny(route.Path, "?#") {
				validationErrors = append(validationErrors,
					fmt.Errorf("routes[%d].path: must not contain query or fragment", i))
			}
		}

		if route.Match != "exact" && route.Match != "prefix" {
			validationErrors = append(validationErrors,
				fmt.Errorf("routes[%d].match: unsupported value %q", i, route.Match))
		}

		if _, exists := cfg.Services[route.Service]; !exists {
			validationErrors = append(validationErrors,
				fmt.Errorf("routes[%d].service: unknown service %q",
					i, route.Service))
		}

		// Reject duplicate route definitions.
		routeKey := route.Match + ":" + route.Path

		if _, exists := seenRoutes[routeKey]; exists {
			validationErrors = append(validationErrors,
				fmt.Errorf("routes[%d]: duplicate route definition (%s, %s)",
					i, route.Match, route.Path))
		} else {
			seenRoutes[routeKey] = struct{}{}
		}
	}

	if len(validationErrors) > 0 {
		return fmt.Errorf("configuration validation failed:\n%w",
			errors.Join(validationErrors...))
	}

	return nil
}
