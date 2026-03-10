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
	origCron := os.Getenv("CRON_SCHEDULE")
	defer func() {
		_ = os.Setenv("MINIFLUX_API_URL", origURL)
		_ = os.Setenv("MINIFLUX_API_TOKEN", origToken)
		_ = os.Setenv("PORT", origPort)
		_ = os.Setenv("DAEMON", origDaemon)
		_ = os.Setenv("CRON_SCHEDULE", origCron)
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
		_ = os.Setenv("CRON_SCHEDULE", "*/15 * * * *")

		config, err := LoadConfig()
		require.NoError(t, err)
		assert.Equal(t, "http://miniflux.example.com", config.APIUrl)
		assert.Equal(t, "9090", config.Port)
		assert.True(t, config.Daemon)
		assert.Equal(t, "*/15 * * * *", config.CronSchedule)
	})

	t.Run("Empty cron schedule", func(t *testing.T) {
		_ = os.Setenv("MINIFLUX_API_URL", "http://localhost:8080")
		_ = os.Setenv("MINIFLUX_API_TOKEN", "secret-token")
		_ = os.Setenv("CRON_SCHEDULE", "")

		config, err := LoadConfig()
		require.NoError(t, err)
		assert.Empty(t, config.CronSchedule)
	})

	t.Run("Whitespace-only cron schedule treated as unset", func(t *testing.T) {
		_ = os.Setenv("MINIFLUX_API_URL", "http://localhost:8080")
		_ = os.Setenv("MINIFLUX_API_TOKEN", "secret-token")
		_ = os.Setenv("CRON_SCHEDULE", "   ")

		config, err := LoadConfig()
		require.NoError(t, err)
		assert.Empty(t, config.CronSchedule)
	})

	t.Run("Invalid cron schedule returns error", func(t *testing.T) {
		_ = os.Setenv("MINIFLUX_API_URL", "http://localhost:8080")
		_ = os.Setenv("MINIFLUX_API_TOKEN", "secret-token")
		_ = os.Setenv("CRON_SCHEDULE", "not-a-valid-cron")

		config, err := LoadConfig()
		require.Error(t, err)
		assert.Nil(t, config)
		assert.Contains(t, err.Error(), "invalid CRON_SCHEDULE")
	})
}
