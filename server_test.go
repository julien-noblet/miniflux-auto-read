package main

import (
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

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
	})

	t.Run("NewHTTPServer", func(t *testing.T) {
		mux := http.NewServeMux()
		srv := NewHTTPServer("9000", mux)
		assert.Equal(t, ":9000", srv.Addr)
		assert.Equal(t, 15*time.Second, srv.ReadTimeout)
	})

	t.Run("StartShutdown", func(t *testing.T) {
		mux := http.NewServeMux()
		srv := NewHTTPServer("9001", mux)

		go Start(srv)

		// Give it a moment to start
		time.Sleep(100 * time.Millisecond)

		Shutdown(srv, 5*time.Second)
	})
}
