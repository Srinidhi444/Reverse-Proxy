package proxy

import (
	"log/slog"
	"net/http"
	"net/http/httputil"

	"github.com/Srinidhi444/Reverse-Proxy/internal/upstream"
)

type Proxy struct {
	logger *slog.Logger
}

func New(logger *slog.Logger) *Proxy {
	return &Proxy{
		logger: logger,
	}
}

func (p *Proxy) Forward(
	w http.ResponseWriter,
	r *http.Request,
	endpoint *upstream.Endpoint,
	transport http.RoundTripper,
) {
	targetURL := endpoint.URL

	reverseProxy := &httputil.ReverseProxy{
		Transport: transport,

		Rewrite: func(pr *httputil.ProxyRequest) {
			pr.SetURL(targetURL)
			pr.SetXForwarded()
		},

		ErrorHandler: func(w http.ResponseWriter, r *http.Request, err error) {
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

	reverseProxy.ServeHTTP(w, r)
}
