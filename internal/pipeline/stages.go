package pipeline

import (
	"log/slog"

	"github.com/trancee/DealScout/internal/currency"
	"github.com/trancee/DealScout/internal/deal"
	"github.com/trancee/DealScout/internal/parser"
	"github.com/trancee/DealScout/internal/parser/cleaners"
)

// transformProduct cleans, normalizes, filters, divides price, and converts currency.
// Returns the cleaned name, CHF price, divided old price, and whether the product should be skipped.
func transformProduct(p parser.RawProduct, category string, priceDivisor float64, priceCurrency string, shopClean cleaners.CleanFunc, catFilter cleaners.FilterFunc, conv *currency.Converter) (string, float64, *float64, bool) {
	var oldPrice *float64
	if priceDivisor > 0 {
		p.Price /= priceDivisor
		if p.OldPrice != nil {
			divided := *p.OldPrice / priceDivisor
			oldPrice = &divided
		}
	} else {
		oldPrice = p.OldPrice
	}

	cleaned := p.Title
	if shopClean != nil {
		cleaned = shopClean(cleaned)
	}
	cleaned = cleaners.NormalizeName(cleaned, category)

	if catFilter != nil && catFilter(cleaned) {
		return "", 0, nil, true
	}

	priceCHF, err := conv.Convert(p.Price, priceCurrency)
	if err != nil {
		slog.Warn("currency conversion failed", "product", cleaned, "error", err)
		return "", 0, nil, true
	}

	return cleaned, priceCHF, oldPrice, false
}

// evaluateProduct runs deal evaluation and builds a ProductResult.
func evaluateProduct(cleaned string, priceCHF float64, oldPrice *float64, p parser.RawProduct, category, shopName string, eval *deal.Evaluator, seedMode bool) (ProductResult, *deal.Deal) {
	result := eval.Evaluate(cleaned, category, shopName, priceCHF, oldPrice, p.URL, p.ImageURL)

	pr := ProductResult{
		Name:     cleaned,
		Shop:     shopName,
		Price:    priceCHF,
		OldPrice: oldPrice,
		URL:      p.URL,
		Reason:   result.Reason,
	}

	if !seedMode && result.Deal != nil {
		pr.IsDeal = true
		pr.Discount = result.Deal.DiscountPct
		return pr, result.Deal
	}

	return pr, nil
}
