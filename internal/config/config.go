package config

import (
	"flag"
)

// Config holds application configuration
type Config struct {
	Port         string
	TelegramToken string // Telegram Bot API token
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

	fs := flag.NewFlagSet("app", flag.ContinueOnError)

	port := fs.String("port", getEnv("APP_PORT", "8000"), "Port to listen on")
	telegramToken := fs.String("telegram-token", getEnv("TELEGRAM_BOT_TOKEN", ""), "Telegram Bot API token")
	// Add your application-specific flags here
	// Example:
	// databaseURL := fs.String("database-url", getEnv("DATABASE_URL", ""), "Database connection URL")
	// apiKey := fs.String("api-key", getEnv("API_KEY", ""), "API key")

	if err := fs.Parse(args[1:]); err != nil {
		return nil, err
	}

	return &Config{
		Port:          *port,
		TelegramToken: *telegramToken,
		// Add your application-specific config assignments here
		// Example:
		// DatabaseURL: *databaseURL,
		// APIKey:      *apiKey,
	}, nil
}
