package router

import (
	"errors"
	"net/http"

	"github.com/Srinidhi444/Reverse-Proxy/internal/config"
	"github.com/Srinidhi444/Reverse-Proxy/internal/upstream"
)

var ErrNoRoute = errors.New("no matching route found")

type Router struct {
	routes   []config.RouteConfig
	services map[string]*upstream.Service
}

func New(routes []config.RouteConfig, services map[string]*upstream.Service) *Router {
	return &Router{
		routes:   routes,
		services: services,
	}
}

func (r *Router) Match(req *http.Request) (*upstream.Service, error) {
	path := req.URL.Path

	// 1. Exact routes take precedence.
	for _, route := range r.routes {
		if route.Match == "exact" && path == route.Path {
			return r.services[route.Service], nil
		}
	}

	// 2. Find the longest matching prefix.
	bestRouteIndex := -1
	longestPrefix := ""

	for i, route := range r.routes {
		if route.Match == "prefix" && matchesPrefix(path, route.Path) {
			if len(route.Path) > len(longestPrefix) {
				longestPrefix = route.Path
				bestRouteIndex = i
			}
		}
	}

	if bestRouteIndex != -1 {
		route := r.routes[bestRouteIndex]
		return r.services[route.Service], nil
	}

	return nil, ErrNoRoute
}

func matchesPrefix(path, prefix string) bool {
	// Exact equality is also a prefix match.
	if path == prefix {
		return true
	}

	// The request path must start with the configured prefix.
	if len(path) < len(prefix) || path[:len(prefix)] != prefix {
		return false
	}

	// Root prefix matches every absolute path.
	if prefix == "/" {
		return true
	}

	// Enforce path-segment boundaries.
	return path[len(prefix)] == '/'
}
