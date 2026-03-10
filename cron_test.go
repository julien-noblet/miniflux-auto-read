package main

import (
	"testing"
	"time"

	"github.com/robfig/cron/v3"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	c "miniflux.app/v2/client"
)

func TestCronScheduling(t *testing.T) {
	// Setup
	mockClient := new(MockMinifluxClient)
	server := &Server{client: mockClient}

	// Expect at least one call to Entries when the cron job fires.
	mockClient.On("Entries", mock.Anything).Return(&c.EntryResultSet{
		Total:   0,
		Entries: c.Entries{},
	}, nil)

	triggered := make(chan struct{}, 1)

	// Use a seconds-enabled parser so we can schedule the job to fire every second.
	scheduler := cron.New(cron.WithSeconds())
	_, err := scheduler.AddFunc("* * * * * *", func() {
		server.Process(&c.Filter{Status: c.EntryStatusUnread})
		select {
		case triggered <- struct{}{}:
		default:
		}
	})
	require.NoError(t, err)

	scheduler.Start()
	defer scheduler.Stop()

	select {
	case <-triggered:
		// Job fired at least once – cron integration works.
	case <-time.After(5 * time.Second):
		t.Fatal("timed out waiting for cron job to fire")
	}

	mockClient.AssertExpectations(t)
}

func TestCronInvalidSchedule(t *testing.T) {
	scheduler := cron.New()
	_, err := scheduler.AddFunc("invalid-schedule", func() {})
	require.Error(t, err)
}

func TestCronValidSchedule(t *testing.T) {
	scheduler := cron.New()
	// Test standard format
	_, err := scheduler.AddFunc("*/15 * * * *", func() {})
	require.NoError(t, err)

	// Test comma format: 0,5 * * * *
	_, err = scheduler.AddFunc("0,5 * * * *", func() {})
	require.NoError(t, err, "Should support commas in cron schedule")

	// Test increment/step format: 0/10 * * * *
	_, err = scheduler.AddFunc("0/10 * * * *", func() {})
	require.NoError(t, err, "Should support step/increment (/) in cron schedule")
}
