package server

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/Srinidhi444/Reverse-Proxy/internal/proxy"
	"github.com/Srinidhi444/Reverse-Proxy/internal/router"
)

type Server struct {
	httpServer *http.Server
	router     *router.Router
	logger     *slog.Logger
	proxy      *proxy.Proxy
}

const requestTimeout = 30 * time.Second

func New(
	address string,
	readTimeout, writeTimeout, idleTimeout time.Duration,
	router *router.Router,
	logger *slog.Logger,
	proxyManager *proxy.Proxy,
) *Server {
	srv := &Server{
		router: router,
		logger: logger,
		proxy:  proxyManager,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", srv.handleRequest)

	srv.httpServer = &http.Server{
		Addr:         address,
		Handler:      mux,
		ReadTimeout:  readTimeout,
		WriteTimeout: writeTimeout,
		IdleTimeout:  idleTimeout,
	}

	return srv
}

func (s *Server) handleRequest(w http.ResponseWriter, r *http.Request) {
	service, err := s.router.Match(r)
	if err != nil {
		// Existing 404 handling
		http.NotFound(w, r)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), requestTimeout)
	defer cancel()

	r = r.WithContext(ctx)

	lease, err := service.Balancer.Acquire()
	if err != nil {
		// Existing 503 handling
		http.Error(w, "Service unavailable", http.StatusServiceUnavailable)
		return
	}

	defer lease.Release()

	endpoint := lease.Endpoint

	// Existing logging

	s.proxy.Forward(w, r, endpoint)
}

func (s *Server) Start() error {
	s.logger.Info(
		"starting HTTP server",
		"address", s.httpServer.Addr,
	)

	if err := s.httpServer.ListenAndServe(); err != nil &&
		!errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("failed to start server: %w", err)
	}

	return nil
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.httpServer.Shutdown(ctx)
}
