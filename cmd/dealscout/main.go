package main

import (
	"flag"
	"fmt"
	"log/slog"
	"os"

	"github.com/trancee/DealScout/internal/config"
	"github.com/trancee/DealScout/internal/notifier"
	"github.com/trancee/DealScout/internal/pipeline"
	"github.com/trancee/DealScout/internal/storage"
)

func main() {
	configDir := flag.String("config", "./config/", "Path to config directory")
	seed := flag.Bool("seed", false, "Populate DB without sending notifications")
	dryRun := flag.Bool("dry-run", false, "Full pipeline, log deals, don't notify")
	shopFilter := flag.String("shop", "", "Run only the named shop")
	testTelegram := flag.Bool("test-telegram", false, "Send a test message to each Telegram topic and exit")
	flag.Parse()

	cfg, err := config.Load(*configDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "fatal: %v\n", err)
		os.Exit(1)
	}

	initLogger(cfg.Settings.LogLevel, cfg.Settings.LogFormat)

	if *testTelegram {
		n := notifier.New(cfg.Secrets.TelegramBotToken, cfg.Secrets.TelegramChannel, cfg.Settings.TelegramTopics).
			WithProxy(cfg.Settings.Proxy)
		sent, err := n.SendTestMessage()
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
		}
		fmt.Printf("Test messages sent: %d/%d topics\n", sent, max(len(cfg.Settings.TelegramTopics), 1))
		if sent == 0 {
			os.Exit(1)
		}
		return
	}

	slog.Info("config loaded",
		"shops", len(cfg.Shops),
		"deal_rules", len(cfg.DealRules),
		"base_currency", cfg.Settings.BaseCurrency,
	)

	db, err := storage.Open(cfg.Settings.DatabasePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "fatal: %v\n", err)
		os.Exit(1)
	}
	defer func() { _ = db.Close() }()

	pipeline.Run(cfg, db, pipeline.Options{
		Seed:     *seed,
		DryRun:   *dryRun,
		ShopName: *shopFilter,
		DumpDir:  cfg.Settings.DumpDir,
	})
}

func initLogger(level, format string) {
	var lvl slog.Level
	switch level {
	case "DEBUG":
		lvl = slog.LevelDebug
	case "WARNING":
		lvl = slog.LevelWarn
	case "ERROR":
		lvl = slog.LevelError
	default:
		lvl = slog.LevelInfo
	}

	opts := &slog.HandlerOptions{Level: lvl}

	var handler slog.Handler
	if format == "json" {
		handler = slog.NewJSONHandler(os.Stdout, opts)
	} else {
		handler = slog.NewTextHandler(os.Stdout, opts)
	}

	slog.SetDefault(slog.New(handler))
}
