package pipeline

import (
	"encoding/base64"
	"fmt"
	"regexp"
	"strings"

	"github.com/trancee/DealScout/internal/config"
)

var base64PlaceholderRe = regexp.MustCompile(`\{base64_start\}(.*?)\{base64_end\}`)

func resolvePricePlaceholders(cat config.ShopCategory, rules map[string]config.DealRule, priceBuckets *config.PriceBuckets) (config.ShopCategory, map[string]string) {
	replacements := map[string]string{}
	rule, ok := rules[cat.Category]
	if !ok {
		return cat, replacements
	}

	minPrice := fmt.Sprintf("%.0f", rule.MinPrice)
	maxPrice := fmt.Sprintf("%.0f", rule.MaxPrice)

	replacements["{min_price}"] = minPrice
	replacements["{max_price}"] = maxPrice

	// Resolve {price_buckets} — select overlapping buckets and format them.
	if priceBuckets != nil {
		replacements["{price_buckets}"] = buildPriceBuckets(priceBuckets, rule.MinPrice, rule.MaxPrice)
	}

	r := strings.NewReplacer("{min_price}", minPrice, "{max_price}", maxPrice)
	cat.URL = r.Replace(cat.URL)
	for i := range cat.URLs {
		cat.URLs[i] = r.Replace(cat.URLs[i])
	}

	// Resolve {price_buckets} in URLs.
	if buckets, ok := replacements["{price_buckets}"]; ok {
		cat.URL = strings.ReplaceAll(cat.URL, "{price_buckets}", buckets)
		for i := range cat.URLs {
			cat.URLs[i] = strings.ReplaceAll(cat.URLs[i], "{price_buckets}", buckets)
		}
	}

	// Resolve {base64_start}...{base64_end} — encode inner content as base64.
	cat.URL = base64PlaceholderRe.ReplaceAllStringFunc(cat.URL, func(match string) string {
		inner := match[len("{base64_start}") : len(match)-len("{base64_end}")]
		return base64.StdEncoding.EncodeToString([]byte(inner))
	})

	return cat, replacements
}

// buildPriceBuckets selects price buckets that overlap with [minPrice, maxPrice]
// and formats them using the configured format string.
func buildPriceBuckets(pb *config.PriceBuckets, minPrice, maxPrice float64) string {
	var parts []string
	for _, r := range pb.Ranges {
		if r.Start < maxPrice && r.End > minPrice {
			entry := pb.Format
			entry = strings.ReplaceAll(entry, "{start}", formatSwissPrice(r.Start))
			entry = strings.ReplaceAll(entry, "{end}", formatSwissPrice(r.End))
			parts = append(parts, entry)
		}
	}
	return strings.Join(parts, "")
}

// formatSwissPrice formats a number with URL-encoded Swiss-style apostrophe
// thousands separator. Values >= 1000 get the separator
// (e.g. 1000 → "1%27000", 100000 → "100%27000").
func formatSwissPrice(v float64) string {
	if v == float64(int(v)) {
		n := int(v)
		if n < 1000 {
			return fmt.Sprintf("%d", n)
		}
		s := fmt.Sprintf("%d", n)
		var result []byte
		for i, c := range s {
			if i > 0 && (len(s)-i)%3 == 0 {
				result = append(result, '%', '2', '7')
			}
			result = append(result, byte(c))
		}
		return string(result)
	}
	return fmt.Sprintf("%.2f", v)
}
