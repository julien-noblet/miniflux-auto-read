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
	defer func() {
		os.Setenv("MINIFLUX_API_URL", origURL)
		os.Setenv("MINIFLUX_API_TOKEN", origToken)
		os.Setenv("DAEMON", origDaemon)
	}()

	t.Run("Run once mode (no daemon)", func(t *testing.T) {
		os.Setenv("PORT", "9292")
		os.Setenv("MINIFLUX_API_URL", "http://localhost:8080")
		os.Setenv("MINIFLUX_API_TOKEN", "token")
		os.Setenv("DAEMON", "false")

		err := Run()
		assert.NoError(t, err)
	})

	t.Run("Config Error", func(t *testing.T) {
		os.Unsetenv("MINIFLUX_API_URL")
		err := Run()
		assert.Error(t, err)
	})
}
