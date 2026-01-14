package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"

	"github.com/en9inerd/postpal/internal/config"
	"github.com/en9inerd/postpal/internal/git"
	"github.com/en9inerd/postpal/internal/handlers"
	"github.com/en9inerd/postpal/internal/log"
	"github.com/en9inerd/postpal/internal/zola"
	"github.com/en9inerd/postpal/pkg/tgbot"
)

var version = "dev"

func run(ctx context.Context, args []string, getenv func(string) string) error {
	ctx, cancel := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer cancel()

	cleanedArgs, verbose := cleanArgs(args)

	cfg, err := config.ParseConfig(cleanedArgs, getenv)
	if err != nil {
		return fmt.Errorf("failed to parse config: %w", err)
	}

	logger := log.NewLogger(verbose)
	logger.Info("starting postpal", "version", version)

	if cfg.TelegramAPIID == 0 {
		return fmt.Errorf("TELEGRAM_API_ID is required")
	}
	if cfg.TelegramAPIHash == "" {
		return fmt.Errorf("TELEGRAM_API_HASH is required")
	}
	if cfg.TelegramBotToken == "" {
		return fmt.Errorf("TELEGRAM_BOT_TOKEN is required")
	}
	if cfg.Channel == "" {
		return fmt.Errorf("TELEGRAM_CHANNEL or TELEGRAM_CHANNEL_ID is required")
	}
	if cfg.Author == "" {
		return fmt.Errorf("TELEGRAM_AUTHOR or TELEGRAM_AUTHOR_ID is required")
	}
	if cfg.GitRepoURL == "" {
		return fmt.Errorf("GIT_REPO_URL is required")
	}

	bot, err := tgbot.New(tgbot.Config{
		APIID:        cfg.TelegramAPIID,
		APIHash:      cfg.TelegramAPIHash,
		BotToken:     cfg.TelegramBotToken,
		SessionDir:   cfg.SessionDir,
		SyncCommands: true,
		Logger:       logger,
		Verbose:      verbose,
	})
	if err != nil {
		return fmt.Errorf("failed to create bot: %w", err)
	}

	gitSvc := git.NewService(
		cfg.GitRepoDir,
		cfg.GitRepoURL,
		cfg.GitBranch,
		cfg.GitAuthToken,
		git.Author{Name: cfg.GitAuthorName, Email: cfg.GitAuthorEmail},
	)

	bot.OnReady(func(ctx context.Context) {
		logger.Info("bot ready, resolving identifiers")

		channelID, err := resolveID(ctx, bot, cfg.Channel, true)
		if err != nil {
			logger.Error("failed to resolve channel", "channel", cfg.Channel, "error", err)
			return
		}

		authorID, err := resolveID(ctx, bot, cfg.Author, false)
		if err != nil {
			logger.Error("failed to resolve author", "author", cfg.Author, "error", err)
			return
		}

		logger.Info("resolved identifiers", "channel_id", channelID, "author_id", authorID)

		postsDir := filepath.Join(cfg.GitRepoDir, cfg.ZolaPostsDir)
		zolaSvc := zola.NewService(
			postsDir,
			cfg.ZolaPostsDir,
			cfg.GitRepoDir,
			fmt.Sprintf("%d", channelID),
			gitSvc,
		)

		h := handlers.New(bot, gitSvc, zolaSvc, channelID, authorID, logger)
		h.Register()

		if gitSvc.RepoExists() {
			logger.Info("pulling latest changes")
			if err := gitSvc.Pull(ctx); err != nil {
				logger.Warn("failed to pull", "error", err)
			}
		} else {
			logger.Info("cloning repository")
			if err := gitSvc.Clone(ctx); err != nil {
				logger.Error("failed to clone repository", "error", err)
				return
			}
		}

		logger.Info("ready to process channel posts",
			"channel", cfg.Channel,
			"author", cfg.Author)
	})

	logger.Info("starting bot")
	if err := bot.Run(ctx); err != nil {
		return fmt.Errorf("bot error: %w", err)
	}

	return nil
}

func main() {
	ctx := context.Background()
	if err := run(ctx, os.Args, os.Getenv); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func cleanArgs(args []string) (cleanArgs []string, verbose bool) {
	for _, arg := range args {
		if arg == "--verbose" || arg == "-v" {
			verbose = true
		} else {
			cleanArgs = append(cleanArgs, arg)
		}
	}
	return
}

// resolveID resolves a string identifier (numeric ID or @username) to a numeric ID.
func resolveID(ctx context.Context, bot *tgbot.Bot, identifier string, isChannel bool) (int64, error) {
	identifier = strings.TrimSpace(identifier)

	if id, err := strconv.ParseInt(identifier, 10, 64); err == nil {
		return id, nil
	}

	username := strings.TrimPrefix(identifier, "@")

	if isChannel {
		channelID, _, err := bot.ResolveChannel(ctx, username)
		if err != nil {
			return 0, fmt.Errorf("failed to resolve channel @%s: %w", username, err)
		}
		return channelID, nil
	}

	userID, _, err := bot.ResolveUser(ctx, username)
	if err != nil {
		return 0, fmt.Errorf("failed to resolve user @%s: %w", username, err)
	}
	return userID, nil
}
