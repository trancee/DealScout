package parser

import (
	"fmt"
	"strings"

	"github.com/trancee/DealScout/internal/config"
)

// Parse extracts products from raw response data using the shop category's configuration.
// Routes to:
//   - Embedded JSON parser (if JSONSelector + Fields are set)
//   - HTML parser (if Selectors are set)
//   - JSON parser (if Fields are set)
func Parse(shopCat config.ShopCategory, data []byte, baseURL string) ([]RawProduct, error) {
	switch {
	case shopCat.JSONSelector != "" && len(shopCat.Fields) > 0:
		return ParseEmbeddedJSON(data, shopCat.JSONSelector, shopCat.Fields)
	case len(shopCat.Selectors) > 0:
		return ParseHTML(data, shopCat.Selectors, baseURL)
	case len(shopCat.Fields) > 0:
		return ParseJSON(data, shopCat.Fields)
	default:
		return nil, fmt.Errorf("shop category %q has neither selectors nor fields configured", shopCat.Category)
	}
}

// ResolveProductURLs resolves relative product URLs against a base URL.
// If urlTemplate is set, it replaces {id} with the product's URL field value.
func ResolveProductURLs(products []RawProduct, baseURL, urlTemplate string) {
	for i := range products {
		if urlTemplate != "" {
			products[i].URL = strings.ReplaceAll(urlTemplate, "{id}", products[i].URL)
		} else if baseURL != "" {
			products[i].URL = resolveURL(products[i].URL, baseURL)
		}
	}
}
