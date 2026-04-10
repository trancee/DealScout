package pipeline

import (
	"testing"

	"github.com/trancee/DealScout/internal/deal"
)

func TestDeduplicateDeals(t *testing.T) {
	t.Run("keeps cheapest across shops", func(t *testing.T) {
		deals := []deal.Deal{
			{ProductName: "Samsung Galaxy A16", Shop: "Brack", Price: 119},
			{ProductName: "Samsung Galaxy A16", Shop: "Conrad", Price: 109},
			{ProductName: "Samsung Galaxy A16", Shop: "Foletti", Price: 129},
		}
		result := deduplicateDeals(deals)
		if len(result) != 1 {
			t.Fatalf("got %d deals, want 1", len(result))
		}
		if result[0].Shop != "Conrad" || result[0].Price != 109 {
			t.Errorf("got shop=%s price=%.0f, want Conrad 109", result[0].Shop, result[0].Price)
		}
	})

	t.Run("keeps different products", func(t *testing.T) {
		deals := []deal.Deal{
			{ProductName: "Samsung Galaxy A16", Shop: "Brack", Price: 119},
			{ProductName: "Apple iPhone SE", Shop: "Conrad", Price: 399},
		}
		result := deduplicateDeals(deals)
		if len(result) != 2 {
			t.Fatalf("got %d deals, want 2", len(result))
		}
	})

	t.Run("case insensitive matching", func(t *testing.T) {
		deals := []deal.Deal{
			{ProductName: "samsung galaxy a16", Shop: "Brack", Price: 119},
			{ProductName: "Samsung Galaxy A16", Shop: "Conrad", Price: 109},
		}
		result := deduplicateDeals(deals)
		if len(result) != 1 {
			t.Fatalf("got %d deals, want 1", len(result))
		}
	})

	t.Run("empty input", func(t *testing.T) {
		result := deduplicateDeals(nil)
		if len(result) != 0 {
			t.Fatalf("got %d deals, want 0", len(result))
		}
	})
}

func TestMarkCheapest(t *testing.T) {
	t.Run("marks cheapest per product", func(t *testing.T) {
		products := []ProductResult{
			{Name: "Samsung Galaxy A16", Shop: "Brack", Price: 119},
			{Name: "Samsung Galaxy A16", Shop: "Conrad", Price: 109},
			{Name: "Apple iPhone SE", Shop: "Brack", Price: 399},
		}
		markCheapest(products)

		if products[0].IsCheapest {
			t.Error("Brack Samsung should not be cheapest")
		}
		if !products[1].IsCheapest {
			t.Error("Conrad Samsung should be cheapest")
		}
		if !products[2].IsCheapest {
			t.Error("only iPhone entry should be cheapest")
		}
	})

	t.Run("single product marked", func(t *testing.T) {
		products := []ProductResult{
			{Name: "Nokia 225", Shop: "Brack", Price: 49},
		}
		markCheapest(products)
		if !products[0].IsCheapest {
			t.Error("single product should be marked cheapest")
		}
	})

	t.Run("empty input no panic", func(t *testing.T) {
		markCheapest(nil)
	})
}
