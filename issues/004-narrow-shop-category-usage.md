# Refactor: Narrow ShopCategory usage — pass specific sub-structs instead of the full config

## Problem

`config.ShopCategory` is a god-struct with 4 inlined sub-structs (`Fetching`, `Parsing`, `Pricing`, `Output`) totaling 15+ fields. It's passed to nearly every function in the pipeline:

```go
func fetchPage(f *fetcher.Fetcher, shop config.Shop, cat config.ShopCategory, ...) ([]byte, error)
func parsePageProducts(data []byte, cat config.ShopCategory, shop config.Shop, ...) ([]parser.RawProduct, error)
func transformProduct(p parser.RawProduct, cat config.ShopCategory, ...) (string, float64, *float64, bool)
func evaluateProduct(..., cat config.ShopCategory, shop config.Shop, ...) (ProductResult, *deal.Deal)
func enrichPrices(..., cat config.ShopCategory, shop config.Shop, ...) []parser.RawProduct
```

Functions that only need a URL and pagination type still receive the full struct including CSS selectors, field mappings, price divisors, and JSONP callbacks. This makes function signatures opaque — you can't tell from the type what data a function actually reads.

In tests, constructing a `ShopCategory` requires setting irrelevant fields to satisfy the type, even when testing simple URL resolution or price division.

## Proposed Interface

Pass the specific sub-struct each function needs:

```go
// fetchPage only needs fetching config
func fetchPage(f *fetcher.Fetcher, shop config.Shop, fetch config.Fetching, ...) ([]byte, error)

// transformProduct only needs pricing config + category name
func transformProduct(p parser.RawProduct, category string, pricing config.Pricing, ...) (string, float64, *float64, bool)

// evaluateProduct only needs category name + shop name
func evaluateProduct(..., category string, shopName string, ...) (ProductResult, *deal.Deal)
```

Where `processShop` iterates categories, it destructures the `ShopCategory` once:

```go
for _, cat := range shop.Categories {
    fetch := cat.Fetching
    parsing := cat.Parsing
    pricing := cat.Pricing
    // pass only what each function needs
}
```

## Dependency Strategy

**In-process.** This is a pure refactor of parameter types. No new dependencies, no behavioral changes. The sub-structs already exist in `config/types.go` — they're just not used directly in function signatures.

## Testing Strategy

- **New boundary tests**: None strictly needed — this is a signature refactor
- **Old tests to simplify**: `stages_test.go` test cases can construct smaller `config.Fetching` or `config.Pricing` structs instead of full `ShopCategory` objects. This reduces test setup noise.
- **Validation**: All existing tests must pass unchanged (behavior is identical)

## Implementation Recommendations

- Start with the leaf functions (`fetchPage`, `transformProduct`, `evaluateProduct`) that have the narrowest actual config needs
- Leave `parsePageProducts` taking the full `ShopCategory` for now — it genuinely uses fields from multiple sub-structs (parsing + pricing + output)
- The `processShop` loop is the natural place to destructure once and pass narrow types downward
- This pairs well with the ShopProcessor refactor (#001) — the processor struct can hold the full config while methods receive narrow parameters internally
- Don't create new wrapper types — use the existing `config.Fetching`, `config.Parsing`, `config.Pricing`, `config.Output` directly
