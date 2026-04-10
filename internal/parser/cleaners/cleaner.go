package cleaners

import (
	"net/url"
	"strings"
)

// CleanFunc transforms a product name (strip artifacts, normalize).
type CleanFunc func(name string) string

// FilterFunc returns true if a product name should be SKIPPED.
type FilterFunc func(name string) bool

var shopCleaners = map[string]CleanFunc{
	"ackermann":     cleanAckermann,
	"alltron":       cleanAlltron,
	"amazon":        cleanAmazon,
	"brack":         cleanBrack,
	"conforama":     cleanConforama,
	"conrad":        cleanConrad,
	"foletti":       cleanFoletti,
	"galaxus":       cleanGalaxus,
	"interdiscount": cleanInterdiscount,
	"mediamarkt":    cleanMediamarkt,
	"mobilezone":    cleanMobilezone,
	"orderflow":     cleanOrderflow,
	"postshop":      cleanPostShop,
	// "cashconverters": cleanCashConverters,
	// "hopcash":        cleanHopCash,
}

var urlCleaners = map[string]CleanFunc{
	"ackermann": stripQueryParams,
	"amazon":    stripQueryParams,
}

// ShopCleaner returns a cleaning function for the given shop, or nil if none.
func ShopCleaner(shopName string) CleanFunc {
	return shopCleaners[strings.ToLower(shopName)]
}

// URLCleaner returns a URL cleaning function for the given shop, or nil if none.
func URLCleaner(shopName string) CleanFunc {
	return urlCleaners[strings.ToLower(shopName)]
}

// stripQueryParams removes all query parameters from a URL.
func stripQueryParams(rawURL string) string {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return rawURL
	}
	parsed.RawQuery = ""
	return parsed.String()
}

// CategoryCleaner returns a cleaning function for the given category, or nil if none.
func CategoryCleaner(category string) CleanFunc {
	return nil
}
