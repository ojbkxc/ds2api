package app

import (
	"net/http"
	"sync"

	"ds2api/internal/config"
	"ds2api/internal/server"
)

var (
	handlerOnce sync.Once
	handler     http.Handler
	appInstance *server.App
)

// NewHandler creates and returns the application HTTP handler.
// It is safe to call multiple times — initialization runs once.
func NewHandler() http.Handler {
	handlerOnce.Do(func() {
		app, err := server.NewApp()
		if err != nil {
			config.Logger.Error("[app] init failed", "error", err)
			handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				server.WriteUnhandledError(w, err)
			})
			return
		}
		handler = app.Router
		appInstance = app
	})
	return handler
}

// Close shuts down the application gracefully, releasing resources
// like MCP host connections, compressor, etc.
func Close() {
	if appInstance != nil {
		appInstance.Close()
	}
}
