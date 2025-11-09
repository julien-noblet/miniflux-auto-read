package main

import (
	"log"
	"os"
)

const (
	defaultPort = "8080"
)

// Config holds the application configuration
type Config struct {
	APIUrl   string
	APIToken string
	Port     string
	Daemon   bool
}

// LoadConfig loads configuration from environment variables
func LoadConfig() *Config {
	apiURL := os.Getenv("MINIFLUX_API_URL")
	apiToken := os.Getenv("MINIFLUX_API_TOKEN")

	if apiURL == "" || apiToken == "" {
		log.Fatal("MINIFLUX_API_URL and MINIFLUX_API_TOKEN must be set")
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = defaultPort
	}

	log.Printf("Using Miniflux API URL: %s", apiURL)
	log.Println("API token configured")

	return &Config{
		APIUrl:   apiURL,
		APIToken: apiToken,
		Port:     port,
		Daemon:   os.Getenv("DAEMON") == "true",
	}
}
