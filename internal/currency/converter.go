package currency

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"

	"github.com/trancee/DealScout/internal/storage"
)

// Converter fetches and caches exchange rates, converting prices to a base currency.
type Converter struct {
	db           *storage.Database
	providerURL  string
	baseCurrency string
	cacheTTL     int
}

// New creates a Converter.
func New(db *storage.Database, providerURL, baseCurrency string, cacheTTLHours int) *Converter {
	return &Converter{
		db:           db,
		providerURL:  providerURL,
		baseCurrency: baseCurrency,
		cacheTTL:     cacheTTLHours,
	}
}

// Convert converts a price from the given currency to the base currency.
func (c *Converter) Convert(price float64, fromCurrency string) (float64, error) {
	if fromCurrency == c.baseCurrency {
		return price, nil
	}

	rate, fresh, err := c.db.ExchangeRate(fromCurrency, c.cacheTTL)
	if err != nil {
		return 0, fmt.Errorf("querying exchange rate for %s: %w", fromCurrency, err)
	}
	if !fresh {
		return 0, fmt.Errorf("no exchange rate available for %s", fromCurrency)
	}

	// rate = how many units of fromCurrency per 1 base currency
	// So: price_in_base = price / rate
	return price / rate, nil
}

// RefreshRates fetches fresh exchange rates from the provider if the cache is stale.
func (c *Converter) RefreshRates() error {
	if c.providerURL == "" {
		return nil
	}

	// Check if any rate is stale by looking at a common currency.
	// If we have a fresh cache, skip the API call.
	_, fresh, _ := c.db.ExchangeRate("EUR", c.cacheTTL)
	if fresh {
		slog.Debug("exchange rates cache is fresh, skipping API call")
		return nil
	}

	return c.fetchAndCacheRates()
}

type rateResponse struct {
	Rates map[string]float64 `json:"rates"`
}

func (c *Converter) fetchAndCacheRates() error {
	url := fmt.Sprintf("%s/latest?from=%s", c.providerURL, c.baseCurrency)
	resp, err := http.Get(url) //nolint:gosec
	if err != nil {
		return fmt.Errorf("fetching exchange rates: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("exchange rate API returned %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("reading exchange rate response: %w", err)
	}

	var rr rateResponse
	if err := json.Unmarshal(body, &rr); err != nil {
		return fmt.Errorf("parsing exchange rate response: %w", err)
	}

	for curr, rate := range rr.Rates {
		if err := c.db.UpsertExchangeRate(curr, rate); err != nil {
			slog.Warn("failed to cache exchange rate", "currency", curr, "error", err)
		}
	}

	slog.Info("exchange rates refreshed", "currencies", len(rr.Rates))
	return nil
}
