package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/robfig/cron/v3"
	c "miniflux.app/v2/client"
)

func main() {
	if err := Run(); err != nil {
		log.Fatal(err)
	}
}

// Run executes the application logic.
func Run() error {
	// Load configuration
	config, err := LoadConfig()
	if err != nil {
		return err
	}

	// Create server
	server := NewServer(config)

	// Setup routes
	mux := server.SetupRoutes()

	// Create HTTP server
	httpServer := NewHTTPServer(config.Port, mux)

	// Start server in background
	go Start(httpServer)

	// If --daemon isn't set, processEntriesHandler will be called once before shutdown
	if !config.Daemon {
		// Call processEntriesHandler
		server.Process(&c.Filter{
			Status: c.EntryStatusUnread,
		})

		Shutdown(httpServer, 30*time.Second)
		return nil
	}

	// Setup Cron Scheduler if CRON_SCHEDULE is defined
	var scheduler *cron.Cron
	if config.CronSchedule != "" {
		utcLocation := time.UTC
		scheduler = cron.New(cron.WithLocation(utcLocation))
		_, err := scheduler.AddFunc(config.CronSchedule, func() {
			log.Printf("Cron job triggered: %s", config.CronSchedule)
			server.Process(&c.Filter{
				Status: c.EntryStatusUnread,
			})
		})
		if err != nil {
			return fmt.Errorf("failed to add cron job with CRON_SCHEDULE %q: %w", config.CronSchedule, err)
		}
		scheduler.Start()
		log.Printf("Cron scheduler started with schedule: %s", config.CronSchedule)
	}

	// Wait for shutdown signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	// Stop the cron scheduler and wait for any in-flight jobs to finish
	if scheduler != nil {
		ctx := scheduler.Stop()
		<-ctx.Done()
	}

	Shutdown(httpServer, 30*time.Second)
	return nil
}
