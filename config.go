// Package main provides the miniflux-auto-read service.
package main

import (
	"errors"
	"log"
	"os"
	"strings"

	"github.com/robfig/cron/v3"
)

const (
	defaultPort = "8080"
)

// Config holds the application configuration.
type Config struct {
	APIUrl       string
	APIToken     string
	Port         string
	Daemon       bool
	CronSchedule string
}

// LoadConfig loads configuration from environment variables.
func LoadConfig() (*Config, error) {
	apiURL := os.Getenv("MINIFLUX_API_URL")
	apiToken := os.Getenv("MINIFLUX_API_TOKEN")

	if apiURL == "" || apiToken == "" {
		return nil, errors.New("MINIFLUX_API_URL and MINIFLUX_API_TOKEN must be set")
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = defaultPort
	}

	cronSchedule := strings.TrimSpace(os.Getenv("CRON_SCHEDULE"))
	if cronSchedule != "" {
		parser := cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow)
		if _, err := parser.Parse(cronSchedule); err != nil {
			return nil, errors.New("invalid CRON_SCHEDULE: " + err.Error())
		}
	}

	log.Println("API token configured")

	return &Config{
		APIUrl:       apiURL,
		APIToken:     apiToken,
		Port:         port,
		Daemon:       os.Getenv("DAEMON") == "true",
		CronSchedule: cronSchedule,
	}, nil
}
