package config

import (
	"flag"
	"strconv"
)

// Config holds all configuration for the application
type Config struct {
	// Telegram MTProto
	TelegramAPIID    int
	TelegramAPIHash  string
	TelegramBotToken string
	Channel          string // Numeric ID or @username
	Author           string // Numeric ID or @username
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
}

// ParseConfig parses configuration from args and environment variables
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

	fs := flag.NewFlagSet("postpal", flag.ContinueOnError)

	// Telegram MTProto
	telegramAPIID := fs.Int("telegram-api-id", getEnvInt("TELEGRAM_API_ID", 0), "Telegram API ID")
	telegramAPIHash := fs.String("telegram-api-hash", getEnv("TELEGRAM_API_HASH", ""), "Telegram API Hash")
	telegramBotToken := fs.String("telegram-bot-token", getEnv("TELEGRAM_BOT_TOKEN", ""), "Telegram Bot Token")
	channel := fs.String("channel", getEnv("TELEGRAM_CHANNEL", ""), "Channel ID or @username")
	author := fs.String("author", getEnv("TELEGRAM_AUTHOR", ""), "Author ID or @username")
	sessionDir := fs.String("session-dir", getEnv("SESSION_DIR", "./session"), "Session storage directory")

	// Git
	gitRepoURL := fs.String("git-repo-url", getEnv("GIT_REPO_URL", ""), "Git repository URL")
	gitRepoDir := fs.String("git-repo-dir", getEnv("GIT_REPO_DIR", "./repo"), "Local repository directory")
	gitBranch := fs.String("git-branch", getEnv("GIT_BRANCH", "master"), "Git branch")
	gitAuthToken := fs.String("git-auth-token", getEnv("GIT_AUTH_TOKEN", ""), "Git authentication token")
	gitAuthorName := fs.String("git-author-name", getEnv("GIT_AUTHOR_NAME", "PostPal"), "Git author name")
	gitAuthorEmail := fs.String("git-author-email", getEnv("GIT_AUTHOR_EMAIL", "bot@postpal.dev"), "Git author email")

	// Zola
	zolaPostsDir := fs.String("zola-posts-dir", getEnv("ZOLA_POSTS_DIR", "content/posts"), "Zola posts directory (relative to repo)")

	if err := fs.Parse(args[1:]); err != nil {
		return nil, err
	}

	return &Config{
		TelegramAPIID:    *telegramAPIID,
		TelegramAPIHash:  *telegramAPIHash,
		TelegramBotToken: *telegramBotToken,
		Channel:          *channel,
		Author:           *author,
		SessionDir:       *sessionDir,
		GitRepoURL:       *gitRepoURL,
		GitRepoDir:       *gitRepoDir,
		GitBranch:        *gitBranch,
		GitAuthToken:     *gitAuthToken,
		GitAuthorName:    *gitAuthorName,
		GitAuthorEmail:   *gitAuthorEmail,
		ZolaPostsDir:     *zolaPostsDir,
	}, nil
}
