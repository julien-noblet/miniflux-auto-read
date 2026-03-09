package main

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoadConfig(t *testing.T) {
	// Original environment restoration
	origURL := os.Getenv("MINIFLUX_API_URL")
	origToken := os.Getenv("MINIFLUX_API_TOKEN")
	origPort := os.Getenv("PORT")
	origDaemon := os.Getenv("DAEMON")
	defer func() {
		os.Setenv("MINIFLUX_API_URL", origURL)
		os.Setenv("MINIFLUX_API_TOKEN", origToken)
		os.Setenv("PORT", origPort)
		os.Setenv("DAEMON", origDaemon)
	}()

	t.Run("Missing variables", func(t *testing.T) {
		os.Unsetenv("MINIFLUX_API_URL")
		os.Unsetenv("MINIFLUX_API_TOKEN")

		config, err := LoadConfig()
		assert.Error(t, err)
		assert.Nil(t, config)
	})

	t.Run("Valid configuration with defaults", func(t *testing.T) {
		os.Setenv("MINIFLUX_API_URL", "http://localhost:8080")
		os.Setenv("MINIFLUX_API_TOKEN", "secret-token")
		os.Unsetenv("PORT")
		os.Unsetenv("DAEMON")

		config, err := LoadConfig()
		assert.NoError(t, err)
		assert.Equal(t, "http://localhost:8080", config.APIUrl)
		assert.Equal(t, "secret-token", config.APIToken)
		assert.Equal(t, "8080", config.Port)
		assert.False(t, config.Daemon)
	})

	t.Run("Custom port and daemon mode", func(t *testing.T) {
		os.Setenv("MINIFLUX_API_URL", "http://miniflux.example.com")
		os.Setenv("MINIFLUX_API_TOKEN", "another-token")
		os.Setenv("PORT", "9090")
		os.Setenv("DAEMON", "true")

		config, err := LoadConfig()
		assert.NoError(t, err)
		assert.Equal(t, "http://miniflux.example.com", config.APIUrl)
		assert.Equal(t, "9090", config.Port)
		assert.True(t, config.Daemon)
	})
}
