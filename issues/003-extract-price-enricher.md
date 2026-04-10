# Refactor: Extract PriceEnricher with injectable fetch and test coverage

## Problem

`enrich.go` (201 lines) handles the two-step price API pattern (search API returns product IDs, secondary API returns prices). It has zero tests despite being the most complex data transformation in the pipeline:

- `buildPriceRequestBody` constructs POST bodies with `{ids}` and `{articles}` placeholders
- `parsePriceResponse` handles both array and map response shapes, extracts prices via jsonpath, and optionally enriches title/image
- `mergePrices` joins price data back to products by URL/ID

The function `enrichPrices` takes 7 parameters including the full `config.Shop`, `config.ShopCategory`, `fetcher.Fetcher`, `responseCache`, and `dumpDir` — most just to support the caching/dumping side effects. The actual price-fetching, parsing, and merging logic is tangled with I/O concerns.

Current bugs hide easily:
- If `parsePriceResponse` misparses a new API shape, no test catches it
- The `strings.TrimLeft(id, "0")` in ID matching is fragile and undocumented
- The Conrad-style `{articles}` placeholder has hardcoded field names

## Proposed Interface

Extract a `PriceEnricher` that separates data fetching from data transformation:

```go
// FetchFunc abstracts the HTTP call for testing.
type FetchFunc func(url string, body string, headers map[string]string) ([]byte, error)

type PriceEnricher struct {
    api   *config.PriceAPI
    fetch FetchFunc
}

func NewPriceEnricher(api *config.PriceAPI, fetch FetchFunc) *PriceEnricher

// Enrich fetches prices from the secondary API and merges them into products.
func (pe *PriceEnricher) Enrich(products []parser.RawProduct) []parser.RawProduct
```

The pure transformation functions become testable independently:

```go
// Exported for testing
func BuildPriceRequestBody(products []parser.RawProduct, api *config.PriceAPI) (string, error)
func ParsePriceResponse(data []byte, api *config.PriceAPI) (map[string]PriceInfo, error)
func MergePrices(products []parser.RawProduct, prices map[string]PriceInfo) []parser.RawProduct
```

## Dependency Strategy

**In-process** for the transformation logic (parse, merge). **Ports & adapters** for the HTTP fetch — inject a `FetchFunc` that can be a real HTTP call in production or a fixture-returning function in tests.

The caching and dumping concerns stay in the pipeline caller — `parsePageProducts` wraps the enricher with cache/dump logic, keeping the enricher itself pure.

## Testing Strategy

- **New boundary tests**:
  - `TestParsePriceResponse_Array` — Conrad-style array response with `productsPath`, `idPath`, `pricePath`
  - `TestParsePriceResponse_Map` — Alltron-style map response keyed by SKU
  - `TestBuildPriceRequestBody_Template` — verify `{ids}` and `{articles}` placeholder replacement
  - `TestMergePrices` — verify join by ID, missing products dropped, title/image enrichment
  - `TestEnrich_EndToEnd` — with a `FetchFunc` returning fixture JSON, verify the full flow
- **Old tests to delete**: None (no existing tests)
- **Test fixtures**: Create `testdata/conrad_prices.json` and `testdata/alltron_tiles.json` fixture files from real API responses

## Implementation Recommendations

- The enricher should own: request body construction, response parsing, price merging
- It should hide: HTTP details, template loading, response format differences (array vs map)
- It should expose: `Enrich(products) products` as the main entry point, plus the pure transformation functions for targeted testing
- The caller (`parsePageProducts`) should handle caching and dumping — these are pipeline concerns, not enrichment concerns
- Consider making `PriceInfo` a public type so callers can inspect intermediate results in debug scenarios
