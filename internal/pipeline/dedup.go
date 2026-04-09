package pipeline

import (
	"log/slog"
	"strings"

	"github.com/trancee/DealScout/internal/deal"
)

// deduplicateDeals keeps only the cheapest deal per product name.
// If the same product appears from multiple shops, only the lowest-priced one is kept.
func deduplicateDeals(deals []deal.Deal) []deal.Deal {
	if len(deals) == 0 {
		return deals
	}

	cheapest := make(map[string]deal.Deal)

	for _, d := range deals {
		key := strings.ToLower(d.ProductName)
		existing, exists := cheapest[key]
		if !exists || d.Price < existing.Price {
			if exists {
				slog.Debug("cross-shop dedup",
					"product", d.ProductName,
					"keeping", d.Shop,
					"price", d.Price,
					"dropping", existing.Shop,
					"was", existing.Price,
				)
			}
			cheapest[key] = d
		}
	}

	result := make([]deal.Deal, 0, len(cheapest))
	for _, d := range cheapest {
		result = append(result, d)
	}

	if dropped := len(deals) - len(result); dropped > 0 {
		slog.Info("cross-shop dedup", "kept", len(result), "dropped", dropped)
	}

	return result
}

// markCheapest marks the cheapest entry for each product name across all shops.
func markCheapest(products []ProductResult) {
	cheapest := make(map[string]int) // key → index of cheapest

	for i, p := range products {
		key := strings.ToLower(p.Name)
		if idx, exists := cheapest[key]; !exists || p.Price < products[idx].Price {
			cheapest[key] = i
		}
	}

	for _, idx := range cheapest {
		products[idx].IsCheapest = true
	}
}
