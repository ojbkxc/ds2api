package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"ds2api/internal/config"
	"ds2api/internal/server"
)

func main() {
	app, err := server.NewApp()
	if err != nil {
		config.Logger.Error("[main] init failed", "error", err)
		os.Exit(1)
	}

	srv := &http.Server{
		Addr:    ":5001",
		Handler: app.Router,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			config.Logger.Error("[main] listen failed", "error", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		config.Logger.Error("[main] shutdown failed", "error", err)
		os.Exit(1)
	}
}
