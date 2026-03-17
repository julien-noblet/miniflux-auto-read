package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestResponseWriter(t *testing.T) {
	t.Run("CaptureStatusCode", func(t *testing.T) {
		rec := httptest.NewRecorder()
		rw := NewResponseWriter(rec)

		rw.WriteHeader(http.StatusNotFound)
		_, _ = rw.Write([]byte("not found"))

		assert.Equal(t, http.StatusNotFound, rw.statusCode)
		assert.Equal(t, http.StatusNotFound, rec.Code)
		assert.Equal(t, "not found", rec.Body.String())
	})
}

func TestPrometheusMiddleware(t *testing.T) {
	t.Run("MetricsRecorded", func(t *testing.T) {
		handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusAccepted)
		})

		mw := prometheusMiddleware(handler)
		req := httptest.NewRequest("GET", "/test", nil)
		rec := httptest.NewRecorder()

		mw.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusAccepted, rec.Code)
	})
}

func TestServerInternal(t *testing.T) {
	config := &Config{
		APIUrl:   "http://localhost:8080",
		APIToken: "token",
		Port:     "8080",
	}

	t.Run("NewServer", func(t *testing.T) {
		s := NewServer(config)
		assert.NotNil(t, s)
		assert.Equal(t, config.APIUrl, s.apiURL)
		assert.Equal(t, config.APIToken, s.apiToken)
		assert.NotNil(t, s.client)
	})

	t.Run("SetupRoutes", func(t *testing.T) {
		s := NewServer(config)
		mux := s.SetupRoutes()
		assert.NotNil(t, mux)

		// Test static endpoints
		endpoints := []string{"/metrics", "/dashboard.json", "/alerts.yaml"}
		for _, ep := range endpoints {
			req := httptest.NewRequest("GET", ep, nil)

			sm, ok := mux.(*http.ServeMux)
			if !ok {
				t.Fatalf("expected *http.ServeMux from SetupRoutes, got %T", mux)
			}

			handler, pattern := sm.Handler(req)
			// We only care that the route is registered; assets may still return 404 at runtime.
			assert.NotNil(t, handler, "Endpoint %s not registered", ep)
			assert.NotEqual(t, "", pattern, "Endpoint %s not registered", ep)
		}
	})

	t.Run("NewHTTPServer", func(t *testing.T) {
		mux := http.NewServeMux()
		srv := NewHTTPServer("9000", mux)
		assert.Equal(t, ":9000", srv.Addr)
		assert.Equal(t, 15*time.Second, srv.ReadTimeout)
	})

	t.Run("StartShutdown", func(_ *testing.T) {
		mux := http.NewServeMux()
		srv := NewHTTPServer("9001", mux)

		go Start(srv)

		// Give it a moment to start
		time.Sleep(100 * time.Millisecond)

		Shutdown(srv, 5*time.Second)
	})
}
