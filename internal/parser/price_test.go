package parser_test

import (
	"testing"

	"github.com/trancee/DealScout/internal/parser"
)

func TestParsePrice(t *testing.T) {
	tests := []struct {
		input string
		want  float64
	}{
		{"CHF 119.–", 119.0},
		{"CHF 119.00", 119.0},
		{"€ 99,90", 99.90},
		{"99,90 €", 99.90},
		{"1'299.00", 1299.0},
		{"1.299,00", 1299.0},
		{"1,299.00", 1299.0},
		{"CHF 1'299.–", 1299.0},
		{"119.00", 119.0},
		{"99.9", 99.9},
		{"EUR 2'499.00", 2499.0},
		{"Fr. 349.–", 349.0},
		{"$199.99", 199.99},
		{"1299", 1299.0},
		{"0.99", 0.99},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := parser.ParsePrice(tt.input)
			if err != nil {
				t.Fatalf("ParsePrice(%q): %v", tt.input, err)
			}
			if got != tt.want {
				t.Errorf("ParsePrice(%q) = %f, want %f", tt.input, got, tt.want)
			}
		})
	}
}

func TestParsePriceErrors(t *testing.T) {
	inputs := []string{"", "free", "N/A", "---"}
	for _, input := range inputs {
		t.Run(input, func(t *testing.T) {
			_, err := parser.ParsePrice(input)
			if err == nil {
				t.Errorf("ParsePrice(%q): expected error", input)
			}
		})
	}
}
