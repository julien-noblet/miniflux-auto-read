package main

import (
	"os"
	"syscall"
	"testing"
	"time"

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

	t.Run("Daemon mode with cron", func(t *testing.T) {
		_ = os.Setenv("PORT", "9293")
		_ = os.Setenv("MINIFLUX_API_URL", "http://localhost:8080")
		_ = os.Setenv("MINIFLUX_API_TOKEN", "token")
		_ = os.Setenv("DAEMON", "true")
		_ = os.Setenv("CRON_SCHEDULE", "0 0 * * *")

		// We can't actually wait for the signal easily without blocking forever,
		// but we can try to run it and send a signal very quickly or just test the setup.
		// Since we want to cover the cron setup, this is enough if we have a way to exit.
		
		go func() {
			time.Sleep(200 * time.Millisecond)
			_ = syscall.Kill(syscall.Getpid(), syscall.SIGTERM)
		}()

		err := Run()
		assert.NoError(t, err)
	})

	t.Run("Invalid Cron Schedule", func(t *testing.T) {
		_ = os.Setenv("DAEMON", "true")
		_ = os.Setenv("CRON_SCHEDULE", "invalid")
		err := Run()
		assert.Error(t, err)
	})
}
