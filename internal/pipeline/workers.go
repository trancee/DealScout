package pipeline

import (
	"bytes"
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/trancee/DealScout/internal/config"
	"github.com/trancee/DealScout/internal/currency"
	"github.com/trancee/DealScout/internal/deal"
	"github.com/trancee/DealScout/internal/fetcher"
	"github.com/trancee/DealScout/internal/parser/cleaners"
)

func collectDeals(shops []config.Shop, f *fetcher.Fetcher, conv *currency.Converter, eval *deal.Evaluator, filters map[string]config.Filter, seedMode bool, dumpDir string, cache *responseCache, summary *Summary) []deal.Deal {
	var (
		mu    sync.Mutex
		deals []deal.Deal
		wg    sync.WaitGroup
		sem   = make(chan struct{}, 5)
	)

	for _, shop := range shops {
		wg.Add(1)
		sem <- struct{}{}

		go func(shop config.Shop) {
			defer wg.Done()
			defer func() { <-sem }()

			clearShopDumpDir(dumpDir, shop.Name)
			sp := NewShopProcessor(shop, f, conv, eval, filters, seedMode, dumpDir, cache)
			result := sp.Process()

			mu.Lock()
			deals = append(deals, result.Deals...)
			summary.Products = append(summary.Products, result.Products...)
			summary.ProductsChecked += result.Count
			summary.Errors += result.Errors
			mu.Unlock()
		}(shop)
	}

	wg.Wait()
	return deals
}

func fetchPage(f *fetcher.Fetcher, shop config.Shop, cat config.ShopCategory, page int, priceReplacements map[string]string) ([]byte, error) {
	switch cat.Pagination.Type {
	case "offset":
		offset := page * cat.Pagination.PerPage
		tpl, err := os.ReadFile(cat.BodyTemplate)
		if err != nil {
			return nil, fmt.Errorf("loading template %s: %w", cat.BodyTemplate, err)
		}
		replacements := map[string]string{"{offset}": fmt.Sprintf("%d", offset)}
		for k, v := range priceReplacements {
			replacements[k] = v
		}
		return f.Post(cat.URL, string(tpl), replacements, shop.Headers)
	case "page_param":
		pageNum := cat.Pagination.Start + page
		url := strings.ReplaceAll(cat.URL, "{page}", fmt.Sprintf("%d", pageNum))
		return f.Get(url, shop.Headers)
	default:
		if cat.BodyTemplate != "" {
			tpl, err := os.ReadFile(cat.BodyTemplate)
			if err != nil {
				return nil, fmt.Errorf("loading template %s: %w", cat.BodyTemplate, err)
			}
			return f.Post(cat.URL, string(tpl), priceReplacements, shop.Headers)
		}
		return f.Get(cat.URL, shop.Headers)
	}
}

func buildFilter(category string, filters map[string]config.Filter) cleaners.FilterFunc {
	f, ok := filters[category]
	if !ok {
		return nil
	}
	return cleaners.NewFilter(f)
}

func stripJSONP(data []byte, callback string) []byte {
	prefix := []byte(callback + "(")
	if !bytes.HasPrefix(data, prefix) {
		return data
	}
	data = data[len(prefix):]
	data = bytes.TrimRight(data, ";\n\r ")
	if len(data) > 0 && data[len(data)-1] == ')' {
		data = data[:len(data)-1]
	}
	return data
}

func categoryURLs(cat config.ShopCategory) []string {
	if len(cat.URLs) > 0 {
		return cat.URLs
	}
	if cat.URL != "" {
		return []string{cat.URL}
	}
	return nil
}

func filterShops(shops []config.Shop, name string) []config.Shop {
	if name == "" {
		return shops
	}
	for _, s := range shops {
		if s.Name == name {
			return []config.Shop{s}
		}
	}
	return nil
}

func fetchMethod(cat config.ShopCategory) string {
	if cat.BodyTemplate != "" || cat.Pagination.Type == "offset" {
		return "POST"
	}
	return "GET"
}

func fetchBody(cat config.ShopCategory, page int, priceReplacements map[string]string) string {
	if cat.BodyTemplate == "" {
		return ""
	}
	tpl, err := os.ReadFile(cat.BodyTemplate)
	if err != nil {
		return ""
	}
	body := string(tpl)
	if cat.Pagination.Type == "offset" {
		offset := page * cat.Pagination.PerPage
		body = strings.ReplaceAll(body, "{offset}", fmt.Sprintf("%d", offset))
	}
	for k, v := range priceReplacements {
		body = strings.ReplaceAll(body, k, v)
	}
	return body
}
