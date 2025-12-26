package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/en9inerd/postpal/internal/config"
	"github.com/en9inerd/postpal/internal/log"
	"github.com/en9inerd/postpal/internal/server"
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
	logger.Info("starting server", "version", version, "port", cfg.Port)

	handler, err := server.NewServer(logger, cfg)
	if err != nil {
		return fmt.Errorf("failed to create server: %w", err)
	}

	httpServer := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      handler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		logger.Info("listening", "addr", httpServer.Addr)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Fprintf(os.Stderr, "error listening and serving: %s\n", err)
		}
	}()

	var wg sync.WaitGroup
	wg.Go(func() {
		<-ctx.Done()
		logger.Info("shutdown signal received")

		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := httpServer.Shutdown(shutdownCtx); err != nil {
			fmt.Fprintf(os.Stderr, "error shutting down http server: %s\n", err)
		}
		logger.Info("server stopped")
	})
	wg.Wait()

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
