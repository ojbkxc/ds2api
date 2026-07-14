package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"ds2api/internal/config"
	"ds2api/internal/server"
)

// shutdownTimeout returns the graceful shutdown duration in seconds.
// Defaults to 120 seconds (enough for SSE streams to finish).
// Configurable via DS2API_SHUTDOWN_TIMEOUT_SECONDS env var.
func shutdownTimeout() time.Duration {
	raw := os.Getenv("DS2API_SHUTDOWN_TIMEOUT_SECONDS")
	if raw == "" {
		return 120 * time.Second
	}
	n, err := strconv.Atoi(raw)
	if err != nil || n <= 0 {
		return 120 * time.Second
	}
	return time.Duration(n) * time.Second
}

func main() {
	app, err := server.NewApp()
	if err != nil {
		config.Logger.Error("[main] init failed", "error", err)
		os.Exit(1)
	}

	srv := &http.Server{
		Addr:              ":5001",
		Handler:           app.Router,
		ReadHeaderTimeout: 10 * time.Second,
		IdleTimeout:       120 * time.Second,
		MaxHeaderBytes:    1 << 20, // 1 MB
	}

	// Register shutdown callback to log and release resources before the server stops.
	srv.RegisterOnShutdown(func() {
		config.Logger.Info("[main] server is shutting down, waiting for active connections to finish...")
	})

	go func() {
		config.Logger.Info("[main] ds2api listening on :5001")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			config.Logger.Error("[main] listen failed", "error", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	sig := <-quit
	timeout := shutdownTimeout()
	config.Logger.Info("[main] received signal, shutting down...", "signal", sig, "timeout", timeout)

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		config.Logger.Error("[main] graceful shutdown failed", "error", err)
		fmt.Fprintf(os.Stderr, "server forced to shutdown: %v\n", err)
		os.Exit(1)
	}

	config.Logger.Info("[main] server exited cleanly")
}
