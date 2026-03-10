package main

import (
	"context"
	"embed"
	"io/fs"
	"log"
	"net/http"
	"strconv"
	"sync/atomic"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	c "miniflux.app/v2/client"
)

//go:embed assets/*.json
var dashboardAssets embed.FS

//go:embed assets/*.yaml
var alertsAssets embed.FS

// ResponseWriter is a wrapper around http.ResponseWriter to capture the status code.
type ResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

// NewResponseWriter creates a new ResponseWriter.
func NewResponseWriter(w http.ResponseWriter) *ResponseWriter {
	return &ResponseWriter{w, http.StatusOK}
}

// WriteHeader captures the status code.
func (rw *ResponseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// prometheusMiddleware measures the duration and status of HTTP requests.
func prometheusMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rw := NewResponseWriter(w)
		next.ServeHTTP(rw, r)

		duration := time.Since(start).Seconds()
		status := strconv.Itoa(rw.statusCode)

		HTTPRequestsTotal.WithLabelValues(r.Method, r.URL.Path, status).Inc()
		HTTPRequestDurationSeconds.WithLabelValues(r.Method, r.URL.Path, status).Observe(duration)
	})
}

// MinifluxClient defines the interface for Miniflux API interactions.
type MinifluxClient interface {
	Me() (*c.User, error)
	Entries(filter *c.Filter) (*c.EntryResultSet, error)
	UpdateEntries(entryIDs []int64, status string) error
	SaveEntry(entryID int64) error
}

// Server represents the HTTP server with Miniflux client.
type Server struct {
	apiURL     string
	apiToken   string
	client     MinifluxClient
	processing atomic.Bool // guards against concurrent Process calls
}

// NewServer creates a new server instance.
func NewServer(config *Config) *Server {
	client := c.NewClient(config.APIUrl, config.APIToken)

	return &Server{
		apiURL:   config.APIUrl,
		apiToken: config.APIToken,
		client:   client,
	}
}

// SetupRoutes configures the HTTP routes.
func (s *Server) SetupRoutes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", s.healthzHandler)
	mux.HandleFunc("/process", s.processEntriesHandler)
	mux.Handle("/metrics", promhttp.Handler())

	// Expose Grafana dashboard JSON
	assets, err := fs.Sub(dashboardAssets, "assets")
	if err == nil {
		mux.Handle("/dashboard.json", http.FileServer(http.FS(assets)))
	}

	// Expose Prometheus alerts YAML
	alerts, err := fs.Sub(alertsAssets, "assets")
	if err == nil {
		mux.Handle("/alerts.yaml", http.FileServer(http.FS(alerts)))
	}

	return prometheusMiddleware(mux)
}

// NewHTTPServer creates and configures the HTTP server.
func NewHTTPServer(port string, handler http.Handler) *http.Server {
	return &http.Server{
		Addr:         ":" + port,
		Handler:      handler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}
}

// Start starts the HTTP server.
func Start(server *http.Server) {
	log.Println("Server starting")
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Server failed to start: %v", err)
	}
}

// Shutdown gracefully shuts down the HTTP server.
func Shutdown(server *http.Server, timeout time.Duration) {
	log.Println("Shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
	}

	log.Println("Server stopped")
}
