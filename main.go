package main

import (
	"os"
	"os/signal"
	"syscall"
	"time"

	c "miniflux.app/v2/client"
)

func main() {
	// Load configuration
	config := LoadConfig()

	// Create server
	server := NewServer(config)

	// Setup routes
	mux := server.SetupRoutes()

	// Create HTTP server
	httpServer := NewHTTPServer(config.Port, mux)

	// Start server in background
	go Start(httpServer)

	// If --daemon isn't set, processEntriesHandler will be called once before shutdown
	if !config.Daemon {

		s := NewServer(config)
		// Call processEntriesHandler
		s.Process(&c.Filter{
			Status: c.EntryStatusUnread,
		})

		Shutdown(httpServer, 30*time.Second)
		return
	}
	// Wait for shutdown signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	Shutdown(httpServer, 30*time.Second)
}
