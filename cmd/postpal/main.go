package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"runtime/debug"
	"syscall"
	"time"

	"github.com/en9inerd/postpal/internal/config"
	"github.com/en9inerd/postpal/internal/git"
	"github.com/en9inerd/postpal/internal/handlers"
	"github.com/en9inerd/postpal/internal/log"
	"github.com/en9inerd/postpal/internal/zola"
	"github.com/en9inerd/telekit"
)

var version = "dev"

func versionString() string {
	var revision, buildTime string
	if info, ok := debug.ReadBuildInfo(); ok {
		for _, kv := range info.Settings {
			switch kv.Key {
			case "vcs.revision":
				if len(kv.Value) >= 7 {
					revision = kv.Value[:7]
				}
			case "vcs.time":
				buildTime = kv.Value
			}
		}
	}
	s := "postpal version " + version
	if revision != "" {
		s += " (" + revision + ")"
	}
	if buildTime != "" {
		s += " built " + buildTime
	}
	return s
}

func run(ctx context.Context, args []string, getenv func(string) string) error {
	for _, a := range args[1:] {
		if a == "--version" || a == "-version" {
			fmt.Println(versionString())
			return nil
		}
	}

	ctx, cancel := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer cancel()

	cfg, err := config.ParseConfig(getenv)
	if err != nil {
		return fmt.Errorf("failed to parse config: %w", err)
	}

	logger := log.NewLogger(cfg.Verbose)
	logger.Info("starting postpal", "version", version)

	bot, err := telekit.New(telekit.Config{
		APIID:        cfg.TelegramAPIID,
		APIHash:      cfg.TelegramAPIHash,
		BotToken:     cfg.TelegramBotToken,
		SessionDir:   cfg.SessionDir,
		Logger:       logger,
		Verbose:      cfg.Verbose,
		SyncCommands: true,
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
		logger,
	)

	bot.OnReady(func(ctx context.Context) {
		logger.Info("bot ready, resolving identifiers")

		channelID, channelAccessHash, channelTitle, err := bot.ResolveIdentifier(ctx, cfg.Channel, true)
		if err != nil {
			logger.Error("failed to resolve channel", "channel", cfg.Channel, "error", err)
			cancel()
			return
		}

		authorID, authorAccessHash, _, err := bot.ResolveIdentifier(ctx, cfg.Author, false)
		if err != nil {
			logger.Error("failed to resolve author", "author", cfg.Author, "error", err)
			cancel()
			return
		}

		logger.Info("resolved identifiers", "channel_id", channelID, "channel_title", channelTitle, "author_id", authorID)

		postsDir := filepath.Join(cfg.GitRepoDir, cfg.ZolaPostsDir)
		zolaSvc := zola.NewService(
			postsDir,
			cfg.ZolaPostsDir,
			cfg.GitRepoDir,
			channelTitle,
			gitSvc,
			logger,
		)

		h := handlers.New(bot, gitSvc, zolaSvc,
			handlers.PeerRef{ID: channelID, AccessHash: channelAccessHash},
			handlers.PeerRef{ID: authorID, AccessHash: authorAccessHash},
			logger,
		)
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
				cancel()
				return
			}
		}

		logger.Info("ready to process channel posts",
			"channel", cfg.Channel,
			"author", cfg.Author)
	})

	// Start health check server
	healthSrv := &http.Server{
		Addr:              ":8080",
		ReadHeaderTimeout: 5 * time.Second,
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
		}),
	}
	go func() {
		if err := healthSrv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("health server error", "error", err)
		}
	}()
	defer func() {
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer shutdownCancel()
		healthSrv.Shutdown(shutdownCtx)
	}()

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
