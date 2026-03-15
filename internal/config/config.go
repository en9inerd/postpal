package config

import (
	"errors"
	"strconv"
)

type Config struct {
	// Telegram MTProto
	TelegramAPIID    int
	TelegramAPIHash  string
	TelegramBotToken string
	Channel          string
	Author           string
	SessionDir       string

	// Git
	GitRepoURL     string
	GitRepoDir     string
	GitBranch      string
	GitAuthToken   string
	GitAuthorName  string
	GitAuthorEmail string

	// Zola
	ZolaPostsDir string

	// Runtime
	Verbose bool
}

func ParseConfig(getenv func(string) string) (*Config, error) {
	cfg := &Config{
		TelegramAPIID:    envInt(getenv, "TELEGRAM_API_ID", 0),
		TelegramAPIHash:  envStr(getenv, "TELEGRAM_API_HASH", ""),
		TelegramBotToken: envStr(getenv, "TELEGRAM_BOT_TOKEN", ""),
		Channel:          envStr(getenv, "TELEGRAM_CHANNEL", ""),
		Author:           envStr(getenv, "TELEGRAM_AUTHOR", ""),
		SessionDir:       envStr(getenv, "SESSION_DIR", "./session"),
		GitRepoURL:       envStr(getenv, "GIT_REPO_URL", ""),
		GitRepoDir:       envStr(getenv, "GIT_REPO_DIR", "./repo"),
		GitBranch:        envStr(getenv, "GIT_BRANCH", "master"),
		GitAuthToken:     envStr(getenv, "GIT_AUTH_TOKEN", ""),
		GitAuthorName:    envStr(getenv, "GIT_AUTHOR_NAME", "PostPal"),
		GitAuthorEmail:   envStr(getenv, "GIT_AUTHOR_EMAIL", "bot@postpal.dev"),
		ZolaPostsDir:     envStr(getenv, "ZOLA_POSTS_DIR", "content/posts"),
		Verbose:          envBool(getenv, "POSTPAL_VERBOSE", false),
	}

	if err := cfg.validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

func envStr(getenv func(string) string, key, fallback string) string {
	if v := getenv(key); v != "" {
		return v
	}
	return fallback
}

func envInt(getenv func(string) string, key string, fallback int) int {
	if v := getenv(key); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			return i
		}
	}
	return fallback
}

func envBool(getenv func(string) string, key string, fallback bool) bool {
	if v := getenv(key); v != "" {
		if b, err := strconv.ParseBool(v); err == nil {
			return b
		}
	}
	return fallback
}

func (c *Config) validate() error {
	if c.TelegramAPIID == 0 {
		return errors.New("TELEGRAM_API_ID is required")
	}
	if c.TelegramAPIHash == "" {
		return errors.New("TELEGRAM_API_HASH is required")
	}
	if c.TelegramBotToken == "" {
		return errors.New("TELEGRAM_BOT_TOKEN is required")
	}
	if c.Channel == "" {
		return errors.New("TELEGRAM_CHANNEL is required")
	}
	if c.Author == "" {
		return errors.New("TELEGRAM_AUTHOR is required")
	}
	if c.GitRepoURL == "" {
		return errors.New("GIT_REPO_URL is required")
	}
	return nil
}
