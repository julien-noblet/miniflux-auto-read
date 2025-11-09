package main

import (
	"os"
	"os/signal"
	"syscall"
	"time"
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

	// Wait for shutdown signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	// Gracefully shutdown the server
	Shutdown(httpServer, 30*time.Second)
}
