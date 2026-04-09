package pipeline_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/trancee/DealScout/internal/config"
	"github.com/trancee/DealScout/internal/pipeline"
	"github.com/trancee/DealScout/internal/storage"
)

func mustOpenDB(t *testing.T) *storage.Database {
	t.Helper()
	db, err := storage.Open(":memory:")
	if err != nil {
		t.Fatalf("Open(:memory:): %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	return db
}

const testHTML = `<html><body>
<div class="product">
  <h2><a href="/p/1"><span>TestPhone 128GB</span></a></h2>
  <span class="price">CHF 199.00</span>
  <img class="img" src="https://img.test/1.jpg">
</div>
<div class="product">
  <h2><a href="/p/2"><span>TestPhone Pro 256GB</span></a></h2>
  <span class="price">CHF 299.00</span>
  <img class="img" src="https://img.test/2.jpg">
</div>
</body></html>`

func TestDryRunFindsDeals(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(testHTML))
	}))
	defer server.Close()

	db := mustOpenDB(t)
	cfg := &config.Config{
		Settings: config.Settings{
			BaseCurrency:              "CHF",
			NotificationCooldownHours: 24,
			FetchDelaySeconds:         0,
			MaxRetries:                1,
			MaxConcurrentShops:        2,
			PriceHistoryRetentionDays: 90,
			DefaultMaxPages:           1,
		},
		Shops: []config.Shop{
			{
				Name:       "TestShop",
				SourceType: "html",
				Categories: []config.ShopCategory{
					{
						Category: "smartphone",
						Fetching: config.Fetching{URL: server.URL, MaxPages: 1},
						Parsing: config.Parsing{
							Selectors: map[string]string{
								"product_card": "div.product",
								"title":        "h2 a span",
								"price":        "span.price",
								"url":          "h2 a[href]",
								"image":        "img.img[src]",
							},
						},
						Pricing: config.Pricing{Currency: "CHF"},
					},
				},
			},
		},
		DealRules: map[string]config.DealRule{
			"smartphone": {MinPrice: 50, MaxPrice: 350, MinDiscountPct: 10},
		},
		Filters: map[string]config.Filter{},
		Secrets: config.Secrets{TelegramBotToken: "tok", TelegramChannel: "-1"},
	}

	summary := pipeline.Run(cfg, db, pipeline.Options{DryRun: true})

	if summary.ProductsChecked != 2 {
		t.Errorf("ProductsChecked = %d, want 2", summary.ProductsChecked)
	}
	if summary.DealsFound != 2 {
		t.Errorf("DealsFound = %d, want 2", summary.DealsFound)
	}
	if summary.NotificationsSent != 0 {
		t.Errorf("NotificationsSent = %d, want 0 (dry-run)", summary.NotificationsSent)
	}
}

func TestSeedStoresButNoDeals(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(testHTML))
	}))
	defer server.Close()

	db := mustOpenDB(t)
	cfg := &config.Config{
		Settings: config.Settings{
			BaseCurrency:              "CHF",
			NotificationCooldownHours: 24,
			FetchDelaySeconds:         0,
			MaxRetries:                1,
			MaxConcurrentShops:        2,
			PriceHistoryRetentionDays: 90,
		},
		Shops: []config.Shop{
			{
				Name: "TestShop",
				Categories: []config.ShopCategory{
					{
						Category: "smartphone",
						Fetching: config.Fetching{URL: server.URL, MaxPages: 1},
						Parsing: config.Parsing{
							Selectors: map[string]string{
								"product_card": "div.product",
								"title":        "h2 a span",
								"price":        "span.price",
								"url":          "h2 a[href]",
								"image":        "img.img[src]",
							},
						},
						Pricing: config.Pricing{Currency: "CHF"},
					},
				},
			},
		},
		DealRules: map[string]config.DealRule{
			"smartphone": {MinPrice: 50, MaxPrice: 350, MinDiscountPct: 10},
		},
		Filters: map[string]config.Filter{},
		Secrets: config.Secrets{TelegramBotToken: "tok", TelegramChannel: "-1"},
	}

	summary := pipeline.Run(cfg, db, pipeline.Options{Seed: true})

	if summary.DealsFound != 0 {
		t.Errorf("DealsFound = %d, want 0 (seed mode)", summary.DealsFound)
	}
	if summary.ProductsChecked != 2 {
		t.Errorf("ProductsChecked = %d, want 2", summary.ProductsChecked)
	}
}

func TestShopFilter(t *testing.T) {
	calls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		_, _ = w.Write([]byte(testHTML))
	}))
	defer server.Close()

	db := mustOpenDB(t)
	cat := config.ShopCategory{
		Category: "smartphone",
		Fetching: config.Fetching{URL: server.URL, MaxPages: 1},
		Parsing:  config.Parsing{Selectors: map[string]string{"product_card": "div.product", "title": "h2 a span", "price": "span.price"}},
		Pricing:  config.Pricing{Currency: "CHF"},
	}
	cfg := &config.Config{
		Settings: config.Settings{BaseCurrency: "CHF", FetchDelaySeconds: 0, MaxRetries: 1, MaxConcurrentShops: 2, PriceHistoryRetentionDays: 90},
		Shops: []config.Shop{
			{Name: "ShopA", Categories: []config.ShopCategory{cat}},
			{Name: "ShopB", Categories: []config.ShopCategory{cat}},
		},
		DealRules: map[string]config.DealRule{"smartphone": {MinPrice: 50, MaxPrice: 350, MinDiscountPct: 10}},
		Filters:   map[string]config.Filter{},
		Secrets:   config.Secrets{TelegramBotToken: "tok", TelegramChannel: "-1"},
	}

	_ = pipeline.Run(cfg, db, pipeline.Options{DryRun: true, ShopName: "ShopA"})

	if calls != 1 {
		t.Errorf("server calls = %d, want 1 (only ShopA)", calls)
	}
}
