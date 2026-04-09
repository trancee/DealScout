package config_test

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/trancee/DealScout/internal/config"
)

func TestLoadValidConfig(t *testing.T) {
	cfg, err := config.Load("testdata/valid")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Settings
	if cfg.Settings.BaseCurrency != "CHF" {
		t.Errorf("BaseCurrency = %q, want %q", cfg.Settings.BaseCurrency, "CHF")
	}
	if cfg.Settings.MaxRetries != 3 {
		t.Errorf("MaxRetries = %d, want %d", cfg.Settings.MaxRetries, 3)
	}
	if cfg.Settings.FetchDelaySeconds != 2 {
		t.Errorf("FetchDelaySeconds = %d, want %d", cfg.Settings.FetchDelaySeconds, 2)
	}
	wantDBPath := "data/dealscout.db"
	if cfg.Settings.DatabasePath != wantDBPath {
		t.Errorf("DatabasePath = %q, want %q", cfg.Settings.DatabasePath, wantDBPath)
	}

	// Shops
	if len(cfg.Shops) != 2 {
		t.Fatalf("len(Shops) = %d, want 2", len(cfg.Shops))
	}
	if cfg.Shops[0].Name != "Galaxus" {
		t.Errorf("Shops[0].Name = %q, want %q", cfg.Shops[0].Name, "Galaxus")
	}
	if cfg.Shops[0].SourceType != "json" {
		t.Errorf("Shops[0].SourceType = %q, want %q", cfg.Shops[0].SourceType, "json")
	}
	if len(cfg.Shops[0].Categories) != 1 {
		t.Fatalf("len(Shops[0].Categories) = %d, want 1", len(cfg.Shops[0].Categories))
	}
	if cfg.Shops[0].Categories[0].Currency != "CHF" {
		t.Errorf("Shops[0].Categories[0].Currency = %q, want %q", cfg.Shops[0].Categories[0].Currency, "CHF")
	}
	if cfg.Shops[1].Name != "Amazon" {
		t.Errorf("Shops[1].Name = %q, want %q", cfg.Shops[1].Name, "Amazon")
	}
	if cfg.Shops[1].BaseURL != "https://www.amazon.de" {
		t.Errorf("Shops[1].BaseURL = %q, want %q", cfg.Shops[1].BaseURL, "https://www.amazon.de")
	}

	// Deal rules
	if len(cfg.DealRules) != 2 {
		t.Fatalf("len(DealRules) = %d, want 2", len(cfg.DealRules))
	}
	sp := cfg.DealRules["smartphone"]
	if sp.MinPrice != 50 {
		t.Errorf("smartphone MinPrice = %f, want 50", sp.MinPrice)
	}
	if sp.MinDiscountPct != 10 {
		t.Errorf("smartphone MinDiscountPct = %f, want 10", sp.MinDiscountPct)
	}

	// Filters
	if len(cfg.Filters) != 2 {
		t.Fatalf("len(Filters) = %d, want 2", len(cfg.Filters))
	}
	if len(cfg.Filters["smartphone"].SkipBrands) != 2 {
		t.Errorf("smartphone SkipBrands count = %d, want 2", len(cfg.Filters["smartphone"].SkipBrands))
	}

	// Secrets
	if cfg.Secrets.TelegramBotToken != "123456:ABC-DEF-test-token" {
		t.Errorf("TelegramBotToken = %q, want %q", cfg.Secrets.TelegramBotToken, "123456:ABC-DEF-test-token")
	}
	if cfg.Secrets.TelegramChannel != "-1001234567890" {
		t.Errorf("TelegramChannel = %q, want %q", cfg.Secrets.TelegramChannel, "-1001234567890")
	}
}

func TestEnvVarsOverrideSecrets(t *testing.T) {
	t.Setenv("TELEGRAM_BOT_TOKEN", "env-token-override")
	t.Setenv("TELEGRAM_CHANNEL", "-999")

	cfg, err := config.Load("testdata/valid")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.Secrets.TelegramBotToken != "env-token-override" {
		t.Errorf("TelegramBotToken = %q, want %q", cfg.Secrets.TelegramBotToken, "env-token-override")
	}
	if cfg.Secrets.TelegramChannel != "-999" {
		t.Errorf("TelegramChannel = %q, want %q", cfg.Secrets.TelegramChannel, "-999")
	}
}

func TestMissingRequiredConfigFails(t *testing.T) {
	_, err := config.Load("testdata/missing_settings")
	if err == nil {
		t.Fatal("expected error for missing config files, got nil")
	}
	if !strings.Contains(err.Error(), "settings.yaml") {
		t.Errorf("error = %q, want it to mention settings.yaml", err.Error())
	}
}

func TestMissingSecretsNoEnvVarsFails(t *testing.T) {
	// Ensure env vars are not set.
	t.Setenv("TELEGRAM_BOT_TOKEN", "")
	t.Setenv("TELEGRAM_CHANNEL", "")

	_, err := config.Load("testdata/no_secrets")
	if err == nil {
		t.Fatal("expected error for missing secrets, got nil")
	}
	if !strings.Contains(err.Error(), "telegram") && !strings.Contains(err.Error(), "secrets") {
		t.Errorf("error = %q, want it to mention telegram or secrets", err.Error())
	}
}

func TestMalformedYAMLFails(t *testing.T) {
	_, err := config.Load("testdata/malformed")
	if err == nil {
		t.Fatal("expected error for malformed YAML, got nil")
	}
	if !strings.Contains(err.Error(), "settings.yaml") {
		t.Errorf("error = %q, want it to mention settings.yaml", err.Error())
	}
}

func TestDefaultsApplied(t *testing.T) {
	cfg, err := config.Load("testdata/minimal")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.Settings.DefaultMaxPages != 5 {
		t.Errorf("DefaultMaxPages = %d, want 5", cfg.Settings.DefaultMaxPages)
	}
	if cfg.Settings.LogLevel != "INFO" {
		t.Errorf("LogLevel = %q, want %q", cfg.Settings.LogLevel, "INFO")
	}
	if cfg.Settings.LogFormat != "text" {
		t.Errorf("LogFormat = %q, want %q", cfg.Settings.LogFormat, "text")
	}
	if cfg.Settings.MaxRetries != 3 {
		t.Errorf("MaxRetries = %d, want 3", cfg.Settings.MaxRetries)
	}
	if cfg.Settings.MaxConcurrentShops != 5 {
		t.Errorf("MaxConcurrentShops = %d, want 5", cfg.Settings.MaxConcurrentShops)
	}
	if cfg.Settings.PriceHistoryRetentionDays != 90 {
		t.Errorf("PriceHistoryRetentionDays = %d, want 90", cfg.Settings.PriceHistoryRetentionDays)
	}
	if cfg.Settings.NotificationCooldownHours != 24 {
		t.Errorf("NotificationCooldownHours = %d, want 24", cfg.Settings.NotificationCooldownHours)
	}
	if cfg.Settings.FetchDelaySeconds != 2 {
		t.Errorf("FetchDelaySeconds = %d, want 2", cfg.Settings.FetchDelaySeconds)
	}
	if cfg.Settings.ExchangeRateCacheTTLHours != 24 {
		t.Errorf("ExchangeRateCacheTTLHours = %d, want 24", cfg.Settings.ExchangeRateCacheTTLHours)
	}
}

func TestPathResolution(t *testing.T) {
	cfg, err := config.Load("testdata/valid")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Database path stays relative to working directory.
	wantDB := "data/dealscout.db"
	if cfg.Settings.DatabasePath != wantDB {
		t.Errorf("DatabasePath = %q, want %q", cfg.Settings.DatabasePath, wantDB)
	}

	// Body template resolved relative to config dir.
	wantTemplate := filepath.Join("testdata/valid", "templates/galaxus_smartphone.json")
	got := cfg.Shops[0].Categories[0].BodyTemplate
	if got != wantTemplate {
		t.Errorf("BodyTemplate = %q, want %q", got, wantTemplate)
	}
}
