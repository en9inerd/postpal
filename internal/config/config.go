package config

import (
	"flag"
	"strconv"
)

type Config struct {
	Port              string
	TelegramToken     string
	AuthPasswordHash  string
	AuthSessionSecret string
	AuthSessionMaxAge int
}

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
	telegramToken := fs.String("telegram-token", getEnv("TELEGRAM_BOT_TOKEN", ""), "Telegram Bot API token")
	authPasswordHash := fs.String("auth-password-hash", getEnv("AUTH_PASSWORD_HASH", ""), "Argon2id password hash")
	authSessionSecret := fs.String("auth-session-secret", getEnv("AUTH_SESSION_SECRET", ""), "Session secret (base64-encoded, 32+ bytes)")
	authSessionMaxAge := fs.Int("auth-session-max-age", getEnvInt("AUTH_SESSION_MAX_AGE", 86400), "Session duration in seconds")

	if err := fs.Parse(args[1:]); err != nil {
		return nil, err
	}

	return &Config{
		Port:              *port,
		TelegramToken:     *telegramToken,
		AuthPasswordHash:  *authPasswordHash,
		AuthSessionSecret: *authSessionSecret,
		AuthSessionMaxAge: *authSessionMaxAge,
	}, nil
}
