package parser

// RawProduct represents a product extracted from a shop listing before cleaning.
type RawProduct struct {
	Title    string
	Price    float64
	OldPrice *float64
	URL      string
	ImageURL string
}
