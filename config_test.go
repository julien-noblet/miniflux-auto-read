package main

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadConfig(t *testing.T) {
	// Original environment restoration
	origURL := os.Getenv("MINIFLUX_API_URL")
	origToken := os.Getenv("MINIFLUX_API_TOKEN")
	origPort := os.Getenv("PORT")
	origDaemon := os.Getenv("DAEMON")
	defer func() {
		_ = os.Setenv("MINIFLUX_API_URL", origURL)
		_ = os.Setenv("MINIFLUX_API_TOKEN", origToken)
		_ = os.Setenv("PORT", origPort)
		_ = os.Setenv("DAEMON", origDaemon)
	}()

	t.Run("Missing variables", func(t *testing.T) {
		_ = os.Unsetenv("MINIFLUX_API_URL")
		_ = os.Unsetenv("MINIFLUX_API_TOKEN")

		config, err := LoadConfig()
		require.Error(t, err)
		assert.Nil(t, config)
	})

	t.Run("Valid configuration with defaults", func(t *testing.T) {
		_ = os.Setenv("MINIFLUX_API_URL", "http://localhost:8080")
		_ = os.Setenv("MINIFLUX_API_TOKEN", "secret-token")
		_ = os.Unsetenv("PORT")
		_ = os.Unsetenv("DAEMON")

		config, err := LoadConfig()
		require.NoError(t, err)
		assert.Equal(t, "http://localhost:8080", config.APIUrl)
		assert.Equal(t, "secret-token", config.APIToken)
		assert.Equal(t, "8080", config.Port)
		assert.False(t, config.Daemon)
	})

	t.Run("Custom port and daemon mode", func(t *testing.T) {
		_ = os.Setenv("MINIFLUX_API_URL", "http://miniflux.example.com")
		_ = os.Setenv("MINIFLUX_API_TOKEN", "another-token")
		_ = os.Setenv("PORT", "9090")
		_ = os.Setenv("DAEMON", "true")

		config, err := LoadConfig()
		require.NoError(t, err)
		assert.Equal(t, "http://miniflux.example.com", config.APIUrl)
		assert.Equal(t, "9090", config.Port)
		assert.True(t, config.Daemon)
	})
}
