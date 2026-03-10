package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	c "miniflux.app/v2/client"
)

func main() {
	if err := Run(); err != nil {
		log.Fatal(err)
	}
}

// Run executes the application logic.
func Run() error {
	// Load configuration
	config, err := LoadConfig()
	if err != nil {
		return err
	}

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
		// Call processEntriesHandler
		server.Process(&c.Filter{
			Status: c.EntryStatusUnread,
		})

		Shutdown(httpServer, 30*time.Second)
		return nil
	}
	// Wait for shutdown signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	Shutdown(httpServer, 30*time.Second)
	return nil
}
