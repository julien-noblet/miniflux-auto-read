package main

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRun(t *testing.T) {
	origURL := os.Getenv("MINIFLUX_API_URL")
	origToken := os.Getenv("MINIFLUX_API_TOKEN")
	origDaemon := os.Getenv("DAEMON")
	origCron := os.Getenv("CRON_SCHEDULE")
	defer func() {
		_ = os.Setenv("MINIFLUX_API_URL", origURL)
		_ = os.Setenv("MINIFLUX_API_TOKEN", origToken)
		_ = os.Setenv("DAEMON", origDaemon)
		_ = os.Setenv("CRON_SCHEDULE", origCron)
	}()

	t.Run("Run once mode (no daemon)", func(t *testing.T) {
		_ = os.Setenv("PORT", "9292")
		_ = os.Setenv("MINIFLUX_API_URL", "http://localhost:8080")
		_ = os.Setenv("MINIFLUX_API_TOKEN", "token")
		_ = os.Setenv("DAEMON", "false")
		_ = os.Unsetenv("CRON_SCHEDULE")

		err := Run()
		assert.NoError(t, err)
	})

	t.Run("Config Error", func(t *testing.T) {
		_ = os.Unsetenv("MINIFLUX_API_URL")
		err := Run()
		assert.Error(t, err)
	})
}
