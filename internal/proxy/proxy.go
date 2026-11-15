package proxy

import (
	"log/slog"
	"net/http"
	"net/http/httputil"

	"github.com/Srinidhi444/Reverse-Proxy/internal/upstream"
)

type Proxy struct {
	logger  *slog.Logger
	proxies map[string]*httputil.ReverseProxy
}

func New(
	logger *slog.Logger,
	services map[string]*upstream.Service,
) *Proxy {
	p := &Proxy{
		logger:  logger,
		proxies: make(map[string]*httputil.ReverseProxy),
	}

	// Build one ReverseProxy per endpoint.
	for _, service := range services {
		for _, endpoint := range service.Endpoints {
			p.proxies[endpoint.ID] = p.newReverseProxy(
				endpoint,
				service.Transport,
			)
		}
	}

	return p
}

func (p *Proxy) newReverseProxy(
	endpoint *upstream.Endpoint,
	transport http.RoundTripper,
) *httputil.ReverseProxy {
	targetURL := endpoint.URL

	return &httputil.ReverseProxy{
		Transport: transport,

		Rewrite: func(pr *httputil.ProxyRequest) {
			pr.SetURL(targetURL)
			pr.SetXForwarded()
		},

		ErrorHandler: func(
			w http.ResponseWriter,
			r *http.Request,
			err error,
		) {
			p.logger.Error(
				"error forwarding request",
				"error", err,
				"method", r.Method,
				"path", r.URL.Path,
				"target", targetURL.String(),
			)

			http.Error(w, "Bad Gateway", http.StatusBadGateway)
		},
	}
}

func (p *Proxy) Forward(
	w http.ResponseWriter,
	r *http.Request,
	endpoint *upstream.Endpoint,
) {
	reverseProxy, exists := p.proxies[endpoint.ID]
	if !exists {
		p.logger.Error(
			"reverse proxy not found",
			"endpoint", endpoint.ID,
		)

		http.Error(w, "Bad Gateway", http.StatusBadGateway)
		return
	}

	reverseProxy.ServeHTTP(w, r)
}
