package pipeline

import (
	"os"
	"strings"
	"testing"

	"github.com/trancee/DealScout/internal/config"
	"github.com/trancee/DealScout/internal/parser"
)

func TestParsePriceResponse_Array(t *testing.T) {
	// Conrad-style: array of products with id and price paths
	data := []byte(`{
		"response": {
			"products": [
				{"articleId": "123", "offers": {"offer": {"price": {"price": 99.95, "crossedOutPrice": 129.95}}}},
				{"articleId": "456", "offers": {"offer": {"price": {"price": 149.00}}}}
			]
		}
	}`)

	api := &config.PriceAPI{
		ProductsPath: "response.products",
		IDPath:       "articleId",
		PricePath:    "offers.offer.price.price",
		OldPricePath: "offers.offer.price.crossedOutPrice",
	}

	prices, err := parsePriceResponse(data, api)
	if err != nil {
		t.Fatal(err)
	}

	if len(prices) != 2 {
		t.Fatalf("got %d prices, want 2", len(prices))
	}

	p123 := prices["123"]
	if p123.price != 99.95 {
		t.Errorf("price[123] = %.2f, want 99.95", p123.price)
	}
	if p123.oldPrice == nil || *p123.oldPrice != 129.95 {
		t.Errorf("oldPrice[123] = %v, want 129.95", p123.oldPrice)
	}

	p456 := prices["456"]
	if p456.price != 149.00 {
		t.Errorf("price[456] = %.2f, want 149.00", p456.price)
	}
	if p456.oldPrice != nil {
		t.Errorf("oldPrice[456] should be nil, got %v", *p456.oldPrice)
	}
}

func TestParsePriceResponse_Map(t *testing.T) {
	// Alltron-style: map keyed by SKU
	data := []byte(`{
		"SKU001": {"sku": "SKU001", "effectivePricing": {"userPrice": 599.00, "mainPrice": 699.00}, "description": {"title": "Laptop X"}, "cover": {"sizes": [{"link": "https://img.jpg"}]}},
		"SKU002": {"sku": "SKU002", "effectivePricing": {"userPrice": 399.00}, "description": {"title": "Laptop Y"}}
	}`)

	api := &config.PriceAPI{
		PricePath:    "effectivePricing.userPrice",
		OldPricePath: "effectivePricing.mainPrice",
		TitlePath:    "description.title",
		ImagePath:    "cover.sizes.0.link",
	}

	prices, err := parsePriceResponse(data, api)
	if err != nil {
		t.Fatal(err)
	}

	if len(prices) != 2 {
		t.Fatalf("got %d prices, want 2", len(prices))
	}

	p1 := prices["SKU001"]
	if p1.price != 599.00 {
		t.Errorf("price = %.2f, want 599.00", p1.price)
	}
	if p1.title != "Laptop X" {
		t.Errorf("title = %q, want Laptop X", p1.title)
	}
	if p1.imageURL != "https://img.jpg" {
		t.Errorf("imageURL = %q", p1.imageURL)
	}
}

func TestMergePrices(t *testing.T) {
	products := []parser.RawProduct{
		{Title: "", URL: "123", Price: 0},
		{Title: "", URL: "456", Price: 0},
		{Title: "", URL: "789", Price: 0}, // no match
	}

	prices := map[string]priceInfo{
		"123": {price: 99.95, title: "Product A", imageURL: "https://a.jpg"},
		"456": {price: 149.00},
	}

	result := mergePrices(products, prices)

	if len(result) != 2 {
		t.Fatalf("got %d products, want 2 (unmatched dropped)", len(result))
	}

	if result[0].Price != 99.95 || result[0].Title != "Product A" || result[0].ImageURL != "https://a.jpg" {
		t.Errorf("product[0] = %+v", result[0])
	}
	if result[1].Price != 149.00 {
		t.Errorf("product[1].Price = %.2f, want 149.00", result[1].Price)
	}
}

func TestBuildPriceRequestBody(t *testing.T) {
	// Create a temp template file
	dir := t.TempDir()
	tplPath := dir + "/template.json"
	tpl := `{"ids": "{ids}", "articles": {articles}}`
	if err := os.WriteFile(tplPath, []byte(tpl), 0o644); err != nil {
		t.Fatal(err)
	}

	products := []parser.RawProduct{
		{URL: "100"},
		{URL: "200"},
	}

	api := &config.PriceAPI{BodyTemplate: tplPath}
	body, err := buildPriceRequestBody(products, api)
	if err != nil {
		t.Fatal(err)
	}

	if !contains(body, "100,200") {
		t.Errorf("body should contain '100,200', got: %s", body)
	}
	if !contains(body, `"articleID":"100"`) {
		t.Errorf("body should contain article 100, got: %s", body)
	}
}

func contains(s, substr string) bool {
	return strings.Contains(s, substr)
}
