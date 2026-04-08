package parser_test

import (
	"os"
	"testing"

	"github.com/trancee/DealScout/internal/parser"
)

func TestParseHTMLExtractsProducts(t *testing.T) {
	html, err := os.ReadFile("testdata/amazon_listing.html")
	if err != nil {
		t.Fatalf("reading fixture: %v", err)
	}

	selectors := map[string]string{
		"product_card": "div[data-component-type='s-search-result']",
		"title":        "h2 a span",
		"price":        "span.a-price span.a-offscreen",
		"old_price":    "span.a-price[data-a-strike] span.a-offscreen",
		"url":          "h2 a[href]",
		"image":        "img.s-image[src]",
	}

	products, err := parser.ParseHTML(html, selectors, "https://www.amazon.de")
	if err != nil {
		t.Fatalf("ParseHTML: %v", err)
	}

	if len(products) != 2 {
		t.Fatalf("len(products) = %d, want 2", len(products))
	}

	// First product.
	p := products[0]
	if p.Title != "Samsung Galaxy A15 128GB Black" {
		t.Errorf("Title = %q, want %q", p.Title, "Samsung Galaxy A15 128GB Black")
	}
	if p.Price != 149.0 {
		t.Errorf("Price = %f, want 149.0", p.Price)
	}
	if p.OldPrice == nil || *p.OldPrice != 179.0 {
		t.Errorf("OldPrice = %v, want 179.0", p.OldPrice)
	}
	if p.URL != "https://www.amazon.de/dp/B0CX1GKQN8" {
		t.Errorf("URL = %q, want %q", p.URL, "https://www.amazon.de/dp/B0CX1GKQN8")
	}
	if p.ImageURL != "https://images.example.com/galaxy-a15.jpg" {
		t.Errorf("ImageURL = %q", p.ImageURL)
	}

	// Second product — no old_price.
	p2 := products[1]
	if p2.Title != "iPhone 16 Pro 256GB" {
		t.Errorf("Title = %q", p2.Title)
	}
	if p2.Price != 1199.0 {
		t.Errorf("Price = %f, want 1199.0", p2.Price)
	}
	if p2.OldPrice != nil {
		t.Errorf("OldPrice should be nil, got %f", *p2.OldPrice)
	}
}
