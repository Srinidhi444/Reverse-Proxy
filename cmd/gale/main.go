package main

import (
	"flag"
	"fmt"
	"log/slog"
	"os"
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

	// 1. Load configuration
	cfg, err := config.Load(*configPath)
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// 2. Build runtime
	rt, err := appRuntime.Build(cfg)
	if err != nil {
		return fmt.Errorf("failed to build runtime: %w", err)
	}
	defer rt.CloseIdleConnections()

	for name, service := range rt.Services {
		logger.Info(
			"service loaded",
			"name", name,
			"endpoints", len(service.Endpoints),
		)
	}

	// 3. Create router
	requestRouter := router.New(cfg.Routes, rt.Services)

	// 4. Create proxy manager
	proxyManager := proxy.New(logger, rt.Services)

	// 5. Create HTTP server
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

	httpServer := server.New(
		cfg.Server.Address,
		readTimeout,
		writeTimeout,
		idleTimeout,
		requestRouter,
		logger,
		proxyManager,
	)

	// 6. Start server
	if err := httpServer.Start(); err != nil {
		return fmt.Errorf("server failed: %w", err)
	}

	return nil
}
