// Package main provides the miniflux-auto-read service.
package main

import (
	"errors"
	"log"
	"os"
)

const (
	defaultPort = "8080"
)

// Config holds the application configuration.
type Config struct {
	APIUrl   string
	APIToken string
	Port     string
	Daemon   bool
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

	log.Println("API token configured")

	return &Config{
		APIUrl:   apiURL,
		APIToken: apiToken,
		Port:     port,
		Daemon:   os.Getenv("DAEMON") == "true",
	}, nil
}
