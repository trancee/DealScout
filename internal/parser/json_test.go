package parser_test

import (
	"os"
	"testing"

	"github.com/trancee/DealScout/internal/parser"
)

func TestParseJSONExtractsProducts(t *testing.T) {
	data, err := os.ReadFile("testdata/galaxus_response.json")
	if err != nil {
		t.Fatalf("reading fixture: %v", err)
	}

	fields := map[string]string{
		"products":  "data.productType.filterProducts.products.results",
		"title":     "product.name",
		"price":     "offer.price.amountInclusive",
		"old_price": "offer.insteadOfPrice.price.amountInclusive",
		"url":       "product.productId",
		"image":     "product.imageUrl",
	}

	products, err := parser.ParseJSON(data, fields)
	if err != nil {
		t.Fatalf("ParseJSON: %v", err)
	}

	if len(products) != 2 {
		t.Fatalf("len(products) = %d, want 2", len(products))
	}

	p := products[0]
	if p.Title != "Samsung Galaxy S24 FE 128 GB" {
		t.Errorf("Title = %q", p.Title)
	}
	if p.Price != 399.0 {
		t.Errorf("Price = %f, want 399.0", p.Price)
	}
	if p.OldPrice == nil || *p.OldPrice != 449.0 {
		t.Errorf("OldPrice = %v, want 449.0", p.OldPrice)
	}
	if p.URL != "42853311" {
		t.Errorf("URL = %q, want %q", p.URL, "42853311")
	}
	if p.ImageURL != "https://images.example.com/s24fe.jpg" {
		t.Errorf("ImageURL = %q", p.ImageURL)
	}

	// Second product — null insteadOfPrice.
	p2 := products[1]
	if p2.Price != 549.0 {
		t.Errorf("Price = %f, want 549.0", p2.Price)
	}
	if p2.OldPrice != nil {
		t.Errorf("OldPrice should be nil, got %f", *p2.OldPrice)
	}
}
