package pipeline

import (
	"testing"

	"github.com/trancee/DealScout/internal/config"
)

func TestBuildPriceBuckets(t *testing.T) {
	buckets := &config.PriceBuckets{
		Format: ":price:CHF+{start}-CHF+{end}",
		Ranges: []config.PriceRange{
			{Start: 0, End: 99.99},
			{Start: 100, End: 199.99},
			{Start: 200, End: 299.99},
			{Start: 300, End: 399.99},
			{Start: 500, End: 599.99},
			{Start: 1000, End: 100000},
		},
	}

	tests := []struct {
		name     string
		min, max float64
		want     string
	}{
		{"selects overlapping buckets", 50, 150,
			":price:CHF+0-CHF+99.99:price:CHF+100-CHF+199.99"},
		{"single bucket", 0, 50,
			":price:CHF+0-CHF+99.99"},
		{"no overlap", 400, 499,
			""},
		{"spans all", 0, 200000,
			":price:CHF+0-CHF+99.99:price:CHF+100-CHF+199.99:price:CHF+200-CHF+299.99:price:CHF+300-CHF+399.99:price:CHF+500-CHF+599.99:price:CHF+1%27000-CHF+100%27000"},
		{"exact boundary", 100, 200,
			":price:CHF+100-CHF+199.99"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := buildPriceBuckets(buckets, tt.min, tt.max)
			if got != tt.want {
				t.Errorf("buildPriceBuckets(%.0f, %.0f) =\n  %q\nwant\n  %q", tt.min, tt.max, got, tt.want)
			}
		})
	}
}

func TestBuildPriceBuckets_Nil(t *testing.T) {
	got := buildPriceBuckets(nil, 50, 150)
	if got != "" {
		t.Errorf("expected empty string for nil buckets, got %q", got)
	}
}

func TestFormatSwissPrice(t *testing.T) {
	tests := []struct {
		input float64
		want  string
	}{
		{0, "0"},
		{99, "99"},
		{500, "500"},
		{999, "999"},
		{1000, "1%27000"},
		{10000, "10%27000"},
		{100000, "100%27000"},
		{1000000, "1%27000%27000"},
		{99.99, "99.99"},
		{599.99, "599.99"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := formatSwissPrice(tt.input)
			if got != tt.want {
				t.Errorf("formatSwissPrice(%v) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestResolvePricePlaceholders(t *testing.T) {
	rules := map[string]config.DealRule{
		"smartphones": {MinPrice: 50, MaxPrice: 150, MinDiscountPct: 10},
	}

	t.Run("replaces min and max", func(t *testing.T) {
		cat := config.ShopCategory{Category: "smartphones"}
		cat.URL = "https://example.com?min={min_price}&max={max_price}"
		cat, replacements := resolvePricePlaceholders(cat, rules, nil)

		if cat.URL != "https://example.com?min=50&max=150" {
			t.Errorf("URL = %q", cat.URL)
		}
		if replacements["{min_price}"] != "50" || replacements["{max_price}"] != "150" {
			t.Errorf("replacements = %v", replacements)
		}
	})

	t.Run("replaces price_buckets", func(t *testing.T) {
		buckets := &config.PriceBuckets{
			Format: ":p:{start}-{end}",
			Ranges: []config.PriceRange{
				{Start: 0, End: 99.99},
				{Start: 100, End: 199.99},
				{Start: 200, End: 299.99},
			},
		}
		cat := config.ShopCategory{Category: "smartphones"}
		cat.URL = "https://example.com?q=test{price_buckets}"
		cat, _ = resolvePricePlaceholders(cat, rules, buckets)

		want := "https://example.com?q=test:p:0-99.99:p:100-199.99"
		if cat.URL != want {
			t.Errorf("URL = %q, want %q", cat.URL, want)
		}
	})

	t.Run("no rule returns unchanged", func(t *testing.T) {
		cat := config.ShopCategory{Category: "monitors"}
		cat.URL = "https://example.com?min={min_price}"
		cat, _ = resolvePricePlaceholders(cat, rules, nil)

		if cat.URL != "https://example.com?min={min_price}" {
			t.Errorf("URL should be unchanged, got %q", cat.URL)
		}
	})
}
