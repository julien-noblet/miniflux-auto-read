package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRun(t *testing.T) {
	t.Run("Run once mode (no daemon)", func(t *testing.T) {
		t.Setenv("PORT", "9292")
		t.Setenv("MINIFLUX_API_URL", "http://localhost:8080")
		t.Setenv("MINIFLUX_API_TOKEN", "token")
		t.Setenv("DAEMON", "false")
		t.Setenv("CRON_SCHEDULE", "")

		err := Run()
		assert.NoError(t, err)
	})

	t.Run("Daemon mode with cron", func(t *testing.T) {
		// This mode currently relies on process-wide OS signals to stop the daemon,
		// which would require sending SIGTERM to the entire 'go test' process.
		// To avoid flakiness and interference with other tests, we skip this test
		// until Run is refactored to accept an injected stop mechanism (e.g. context).
		t.Skip("Skipping daemon mode test to avoid sending SIGTERM to the go test process")
	})

	t.Run("Invalid Cron Schedule", func(t *testing.T) {
		t.Setenv("DAEMON", "true")
		t.Setenv("CRON_SCHEDULE", "invalid")

		err := Run()
		assert.Error(t, err)
	})
}
