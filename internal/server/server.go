package server

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/Srinidhi444/Reverse-Proxy/internal/router"
)

type Server struct {
	httpServer *http.Server
	router     *router.Router
	logger     *slog.Logger
}

func New(
	address string,
	readTimeout, writeTimeout, idleTimeout time.Duration,
	router *router.Router,
	logger *slog.Logger,
) *Server {
	srv := &Server{
		router: router,
		logger: logger,
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
		s.logger.Warn(
			"no route matched",
			"method", r.Method,
			"path", r.URL.Path,
		)

		http.NotFound(w, r)
		return
	}

	endpoint, err := service.Balancer.Next()
	if err != nil {
		s.logger.Error(
			"failed to select endpoint",
			"service", service.Name,
			"error", err,
		)

		http.Error(w, "Service unavailable", http.StatusServiceUnavailable)
		return
	}

	s.logger.Info(
		"endpoint selected",
		"method", r.Method,
		"path", r.URL.Path,
		"service", service.Name,
		"endpoint", endpoint.ID,
	)

	w.WriteHeader(http.StatusOK)
	_, _ = fmt.Fprintf(w, "Selected Endpoint: %s\n", endpoint.URL)
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
