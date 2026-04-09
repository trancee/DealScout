package parser

import (
	"bytes"
	"fmt"
	"net/url"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// ParseHTML extracts products from raw HTML using CSS selectors.
// Required selectors: "product_card", "title", "price".
// Optional selectors: "old_price", "url", "image".
func ParseHTML(html []byte, selectors map[string]string, baseURL string) ([]RawProduct, error) {
	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(html))
	if err != nil {
		return nil, fmt.Errorf("parsing HTML: %w", err)
	}

	cardSel := selectors["product_card"]
	if cardSel == "" {
		return nil, fmt.Errorf("missing required selector: product_card")
	}

	var products []RawProduct

	doc.Find(cardSel).Each(func(_ int, card *goquery.Selection) {
		title := extractField(card, selectors["title"])
		if title == "" {
			return
		}
		if prefix := extractField(card, selectors["title_prefix"]); prefix != "" {
			if !strings.HasPrefix(strings.ToLower(title), strings.ToLower(prefix)) {
				title = prefix + " " + title
			}
		}

		priceStr := extractText(card, selectors["price"])
		price, err := ParsePrice(priceStr)
		if err != nil {
			return
		}

		product := RawProduct{
			Title:    title,
			Price:    price,
			URL:      resolveURL(extractAttr(card, selectors["url"], "href"), baseURL),
			ImageURL: extractAttr(card, selectors["image"], "src"),
		}

		if oldPriceSel := selectors["old_price"]; oldPriceSel != "" {
			if oldPriceStr := extractText(card, oldPriceSel); oldPriceStr != "" {
				if oldPrice, err := ParsePrice(oldPriceStr); err == nil {
					product.OldPrice = &oldPrice
				}
			}
		}

		products = append(products, product)
	})

	return products, nil
}

func extractText(sel *goquery.Selection, cssSelector string) string {
	if cssSelector == "" {
		return ""
	}
	return strings.TrimSpace(sel.Find(cssSelector).First().Text())
}

// extractField extracts text from an element, or an attribute if the selector
// ends with [attr] (e.g., "img.img-fluid[alt]" extracts the alt attribute).
func extractField(sel *goquery.Selection, cssSelector string) string {
	if cssSelector == "" {
		return ""
	}
	if idx := strings.LastIndex(cssSelector, "["); idx > 0 && strings.HasSuffix(cssSelector, "]") {
		attr := cssSelector[idx+1 : len(cssSelector)-1]
		baseSel := cssSelector[:idx]
		val, _ := sel.Find(baseSel).First().Attr(attr)
		return strings.TrimSpace(val)
	}
	return strings.TrimSpace(sel.Find(cssSelector).First().Text())
}

func extractAttr(sel *goquery.Selection, cssSelector, attr string) string {
	if cssSelector == "" {
		return ""
	}
	val, _ := sel.Find(cssSelector).First().Attr(attr)
	return strings.TrimSpace(val)
}

func resolveURL(rawURL, baseURL string) string {
	if rawURL == "" || baseURL == "" {
		return rawURL
	}

	parsed, err := url.Parse(rawURL)
	if err != nil {
		return rawURL
	}

	if parsed.IsAbs() {
		return rawURL
	}

	base, err := url.Parse(baseURL)
	if err != nil {
		return rawURL
	}

	return base.ResolveReference(parsed).String()
}
