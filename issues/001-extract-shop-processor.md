# Refactor: Extract ShopProcessor struct to eliminate parameter threading

## Problem

The `processShop` function in the pipeline package is the central orchestrator for shop processing, but it threads 8 parameters through a deep call chain:

```go
func processShop(shop config.Shop, f *fetcher.Fetcher, conv *currency.Converter,
    eval *deal.Evaluator, filters map[string]config.Filter, seedMode bool,
    dumpDir string, cache *responseCache) ([]deal.Deal, []ProductResult, int, int)
```

Every function it calls (`fetchPageData`, `parsePageProducts`, `enrichPrices`, `transformProduct`, `evaluateProduct`) receives a subset of these same parameters, creating a cascade of 5-8 parameter functions. Adding any new cross-cutting concern (token refresh, rate limiting per category, progress tracking) requires touching every function signature in the chain.

The 4-value return `([]deal.Deal, []ProductResult, int, int)` is another smell — the caller must know that the third `int` is products and the fourth is errors.

## Proposed Interface

Introduce a `ShopProcessor` struct that holds shared state for a single shop run:

```go
type ShopProcessor struct {
    shop      config.Shop
    fetcher   *fetcher.Fetcher
    conv      *currency.Converter
    eval      *deal.Evaluator
    filters   map[string]config.Filter
    seedMode  bool
    cache     *responseCache
    dumpDir   string
    shopClean cleaners.CleanFunc
    urlClean  cleaners.CleanFunc
}

type ShopResult struct {
    Deals    []deal.Deal
    Products []ProductResult
    Count    int
    Errors   int
}

func NewShopProcessor(shop config.Shop, ...) *ShopProcessor
func (sp *ShopProcessor) Process() ShopResult
```

Internal methods become receiver methods with access to shared state:
```go
func (sp *ShopProcessor) fetchPage(cat config.ShopCategory, page int, replacements map[string]string) ([]byte, error)
func (sp *ShopProcessor) parseProducts(data []byte, cat config.ShopCategory) ([]parser.RawProduct, error)
```

## Dependency Strategy

**In-process.** All dependencies are already passed as values/pointers. The struct simply captures them once instead of threading through every call. No new interfaces or adapters needed.

## Testing Strategy

- **New boundary tests**: Test `ShopProcessor.Process()` end-to-end with `httptest` servers providing fixture JSON/HTML responses. Assert on `ShopResult` fields.
- **Old tests to keep**: `stages_test.go` tests for `transformProduct` and `evaluateProduct` remain valid — they test pure logic independent of the processor.
- **Old tests to revisit**: The 3 integration tests in `pipeline_test.go` could be simplified since `ShopProcessor` is easier to construct than calling `processShop` with 8 args.

## Implementation Recommendations

- The `ShopProcessor` should own the full lifecycle of a single shop: token fetch, category iteration, page fetching, parsing, cleaning, evaluation
- It should hide: parameter threading, cache/dump coordination, JSONP stripping, bearer token injection
- It should expose: `Process() ShopResult` as the single entry point
- `collectDeals` becomes a thin concurrent dispatcher that creates `ShopProcessor` instances and merges results
- The struct fields should be private — construction via `NewShopProcessor` ensures all dependencies are provided
