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
	if err := config.Init(); err != nil {
		panic(err)
	}

	app, err := server.NewApp()
	if err != nil {
		config.Logger.Fatal("[main] init failed", "error", err)
	}

	srv := &http.Server{
		Addr:    ":5001",
		Handler: app.Router,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			config.Logger.Fatal("[main] listen failed", "error", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		config.Logger.Fatal("[main] shutdown failed", "error", err)
	}
}