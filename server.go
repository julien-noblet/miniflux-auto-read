package main

import (
	"context"
	"log"
	"net/http"
	"time"

	c "miniflux.app/v2/client"
)

// Server represents the HTTP server with Miniflux client
type Server struct {
	apiURL   string
	apiToken string
	client   *c.Client
}

// NewServer creates a new server instance
func NewServer(config *Config) *Server {
	client := c.NewClient(config.APIUrl, config.APIToken)

	return &Server{
		apiURL:   config.APIUrl,
		apiToken: config.APIToken,
		client:   client,
	}
}

// SetupRoutes configures the HTTP routes
func (s *Server) SetupRoutes() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", s.healthzHandler)
	mux.HandleFunc("/process", s.processEntriesHandler)
	return mux
}

// NewHTTPServer creates and configures the HTTP server
func NewHTTPServer(port string, handler http.Handler) *http.Server {
	return &http.Server{
		Addr:         ":" + port,
		Handler:      handler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}
}

// Start starts the HTTP server
func Start(server *http.Server) {
	log.Printf("Server starting on port %s", server.Addr)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Server failed to start: %v", err)
	}
}

// Shutdown gracefully shuts down the HTTP server
func Shutdown(server *http.Server, timeout time.Duration) {
	log.Println("Shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server stopped")
}
