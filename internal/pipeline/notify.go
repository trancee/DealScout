package pipeline

import (
	"log/slog"
	"time"

	"github.com/trancee/DealScout/internal/config"
	"github.com/trancee/DealScout/internal/deal"
	"github.com/trancee/DealScout/internal/notifier"
	"github.com/trancee/DealScout/internal/storage"
)

func sendNotifications(deals []deal.Deal, cfg *config.Config, db *storage.Database, opts Options, summary *Summary) {
	if opts.Seed || opts.DryRun {
		if opts.DryRun {
			for _, d := range deals {
				slog.Info("deal (dry-run)", "product", d.ProductName, "shop", d.Shop, "price", d.Price, "discount", d.DiscountPct)
			}
		}
		return
	}

	n := notifier.New(cfg.Secrets.TelegramBotToken, cfg.Secrets.TelegramChannel, cfg.Settings.TelegramTopics)
	for _, d := range deals {
		if err := n.Send(d); err != nil {
			slog.Error("notification failed", "product", d.ProductName, "error", err)
			summary.Errors++
			continue
		}
		id, _, _ := db.UpsertProduct(d.ProductName, d.Category)
		if err := db.RecordNotification(id, d.Shop, d.Price); err != nil {
			slog.Error("record notification failed", "product", d.ProductName, "error", err)
		}
		summary.NotificationsSent++
	}
}

func pruneHistory(db *storage.Database, retentionDays int, summary *Summary) {
	deleted, err := db.PruneOldPriceHistory(retentionDays)
	if err != nil {
		slog.Error("prune failed", "error", err)
		summary.Errors++
		return
	}
	if deleted > 0 {
		slog.Info("pruned old price history", "deleted", deleted)
	}
}

func logSummary(s Summary) {
	slog.Info("run complete",
		"products_checked", s.ProductsChecked,
		"deals_found", s.DealsFound,
		"notifications_sent", s.NotificationsSent,
		"errors", s.Errors,
		"duration", s.Duration.Round(time.Millisecond),
	)
}
