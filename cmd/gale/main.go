package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Srinidhi444/Reverse-Proxy/internal/config"
	"github.com/Srinidhi444/Reverse-Proxy/internal/proxy"
	"github.com/Srinidhi444/Reverse-Proxy/internal/router"
	appRuntime "github.com/Srinidhi444/Reverse-Proxy/internal/runtime"
	"github.com/Srinidhi444/Reverse-Proxy/internal/server"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	if err := run(logger); err != nil {
		logger.Error("Gale failed", "error", err)
		os.Exit(1)
	}
}

func run(logger *slog.Logger) error {
	configPath := flag.String(
		"config",
		"configs/gale.yaml",
		"path to config file",
	)
	flag.Parse()

	// 1. Load configuration.
	cfg, err := config.Load(*configPath)
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// 2. Build runtime.
	rt, err := appRuntime.Build(cfg)
	if err != nil {
		return fmt.Errorf("failed to build runtime: %w", err)
	}
	defer rt.CloseIdleConnections()

	// 3. Create application context.
	ctx, cancel := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer cancel()

	// 4. Start health checks and wait for initial checks.
	if err := rt.StartHealthChecks(ctx); err != nil {
		return fmt.Errorf("failed to initialize health checks: %w", err)
	}

	for name, service := range rt.Services {
		logger.Info(
			"service loaded",
			"name", name,
			"endpoints", len(service.Endpoints),
		)
	}

	// 5. Create router.
	requestRouter := router.New(cfg.Routes, rt.Services)

	// 6. Create proxy manager.
	proxyManager := proxy.New(logger, rt.Services)

	// 7. Parse server timeouts.
	readTimeout, err := time.ParseDuration(cfg.Server.ReadTimeout)
	if err != nil {
		return fmt.Errorf("invalid read timeout: %w", err)
	}

	writeTimeout, err := time.ParseDuration(cfg.Server.WriteTimeout)
	if err != nil {
		return fmt.Errorf("invalid write timeout: %w", err)
	}

	idleTimeout, err := time.ParseDuration(cfg.Server.IdleTimeout)
	if err != nil {
		return fmt.Errorf("invalid idle timeout: %w", err)
	}

	// 8. Create HTTP server.
	httpServer := server.New(
		cfg.Server.Address,
		readTimeout,
		writeTimeout,
		idleTimeout,
		requestRouter,
		logger,
		proxyManager,
	)

	// 9. Start HTTP server in the background.
	serverErrors := make(chan error, 1)

	go func() {
		serverErrors <- httpServer.Start()
	}()

	logger.Info("Gale server started", "address", cfg.Server.Address)

	// 10. Wait for either a server error or shutdown signal.
	select {
	case err := <-serverErrors:
		if err != nil {
			return fmt.Errorf("server failed: %w", err)
		}

	case <-ctx.Done():
		logger.Info("shutdown signal received")

		// Use a fresh context because the application context is cancelled.
		shutdownCtx, shutdownCancel := context.WithTimeout(
			context.Background(),
			10*time.Second,
		)
		defer shutdownCancel()

		if err := httpServer.Shutdown(shutdownCtx); err != nil {
			return fmt.Errorf("graceful shutdown failed: %w", err)
		}

		// Wait for the server goroutine to finish.
		if err := <-serverErrors; err != nil {
			return fmt.Errorf("server shutdown failed: %w", err)
		}
	}

	logger.Info("Gale stopped")
	return nil
}
