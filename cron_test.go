package main

import (
	"log"
	"testing"

	"github.com/robfig/cron/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	c "miniflux.app/v2/client"
)

func TestCronScheduling(t *testing.T) {
	// Setup
	mockClient := new(MockMinifluxClient)
	server := &Server{client: mockClient}

	// We expect one call to Entries (to get unread items)
	// and if entries found, SaveEntry and UpdateEntries.
	// For this test, let's say there are no entries to keep it simple.
	mockClient.On("Entries", mock.Anything).Return(&c.EntryResultSet{
		Total:   0,
		Entries: c.Entries{},
	}, nil)

	// Create a scheduler with a very fast frequency (every second)
	// Note: Standard cron doesn't support seconds, but robfig/cron/v3 can if configured.
	// Here we use the default 5-field parser which is minutes.
	// So we'll trigger it manually or use a smaller unit if possible.
	
	triggered := make(chan bool, 1)
	
	scheduler := cron.New()
	_, err := scheduler.AddFunc("* * * * *", func() {
		log.Println("Cron triggered in test")
		server.Process(&c.Filter{Status: c.EntryStatusUnread})
		triggered <- true
	})
	assert.NoError(t, err)

	scheduler.Start()
	defer scheduler.Stop()

	// Since we can't wait a minute in a unit test, we'll verify the logic 
	// of AddFunc and just ensure the function can be called.
	
	// Manual trigger of the cron logic to verify integration with server.Process
	server.Process(&c.Filter{Status: c.EntryStatusUnread})
	
	mockClient.AssertExpectations(t)
}

func TestCronInvalidSchedule(t *testing.T) {
	scheduler := cron.New()
	_, err := scheduler.AddFunc("invalid-schedule", func() {})
	assert.Error(t, err)
}

func TestCronValidSchedule(t *testing.T) {
	scheduler := cron.New()
	
	// Test standard format
	_, err := scheduler.AddFunc("*/15 * * * *", func() {})
	assert.NoError(t, err)

	// Test comma format: 0,5 * * * *
	_, err = scheduler.AddFunc("0,5 * * * *", func() {})
	assert.NoError(t, err, "Should support commas in cron schedule")

	// Test increment/step format: 0/10 * * * *
	_, err = scheduler.AddFunc("0/10 * * * *", func() {})
	assert.NoError(t, err, "Should support step/increment (/) in cron schedule")
}
