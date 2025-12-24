package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"syscall"
	"time"

	"linkedin-automation-poc/internal/app"
	"linkedin-automation-poc/internal/config"
	"linkedin-automation-poc/internal/logger"
	"linkedin-automation-poc/internal/storage"
)

func main() {
	configPath := flag.String("config", "config.yaml", "path to config file")
	dryRun := flag.Bool("dry-run", false, "run local demo mode (no LinkedIn actions)")
	flag.Parse()

	_ = config.LoadDotEnv(".env")

	cfg, err := config.Load(*configPath)
	if err != nil {
		panic(err)
	}

	log, err := logger.New(cfg.Logging)
	if err != nil {
		panic(err)
	}

	store, err := storage.New(cfg.Storage, log)
	if err != nil {
		log.Error("failed to initialize storage", map[string]any{"error": err.Error()})
		os.Exit(1)
	}
	defer func() {
		_ = store.Close()
	}()

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	runner := app.NewRunner(cfg, log, store, *dryRun)
	if err := runner.Run(ctx); err != nil {
		log.Error("run failed", map[string]any{"error": err.Error()})
		time.Sleep(250 * time.Millisecond)
		os.Exit(1)
	}
}
