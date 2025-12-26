package config

import (
	"flag"
	"strconv"
)

// Config holds application configuration
type Config struct {
	Port string
	// Add your application-specific config fields here
	// Example:
	// DatabaseURL string
	// APIKey      string
	// Timeout     time.Duration
}

// ParseConfig parses command-line flags and environment variables
func ParseConfig(args []string, getenv func(string) string) (*Config, error) {
	getEnv := func(key, fallback string) string {
		if v := getenv(key); v != "" {
			return v
		}
		return fallback
	}

	getEnvInt := func(key string, fallback int) int {
		if v := getenv(key); v != "" {
			if i, err := strconv.Atoi(v); err == nil {
				return i
			}
		}
		return fallback
	}

	fs := flag.NewFlagSet("app", flag.ContinueOnError)

	port := fs.String("port", getEnv("APP_PORT", "8000"), "Port to listen on")
	// Add your application-specific flags here
	// Example:
	// databaseURL := fs.String("database-url", getEnv("DATABASE_URL", ""), "Database connection URL")
	// apiKey := fs.String("api-key", getEnv("API_KEY", ""), "API key")

	if err := fs.Parse(args[1:]); err != nil {
		return nil, err
	}

	return &Config{
		Port: *port,
		// Add your application-specific config assignments here
		// Example:
		// DatabaseURL: *databaseURL,
		// APIKey:      *apiKey,
	}, nil
}
