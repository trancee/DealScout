# Add unit tests for untested pipeline infrastructure

## Problem

Five files in the `pipeline` package have zero dedicated tests:

| File | Lines | Responsibility | Risk |
|------|-------|---------------|------|
| `prices.go` | 90 | Price placeholder resolution, bucket filtering, Swiss price formatting | A bug in `buildPriceBuckets` silently fetches wrong products from PostShop |
| `enrich.go` | 201 | Two-step price API (body building, response parsing, merging) | A parsing regression breaks Conrad/Alltron with no test signal |
| `dedup.go` | 66 | Cross-shop deduplication + cheapest marking | A dedup bug causes duplicate Telegram notifications |
| `cache.go` | 80 | File-based TTL response cache | A TTL bug causes stale data or excessive API calls |
| `dump.go` | 90 | Debug response dumps with curl commands | Low risk, but `buildCurlCommand` and `sanitizeName` are easy to unit test |

These are exercised indirectly through 3 integration tests in `pipeline_test.go`, but:
- Integration tests can't isolate which component failed
- Edge cases (empty bucket list, zero-price products, expired cache, map vs array responses) are untested
- The `formatSwissPrice` function added today has no tests for the `%27` encoding logic

## Proposed Test Coverage

### `prices_test.go`

```go
func TestBuildPriceBuckets(t *testing.T)
// - Selects only overlapping ranges (e.g., rule 50-150 selects buckets 0-99.99 + 100-199.99)
// - Returns empty string when no buckets overlap
// - Handles rule range spanning a single bucket
// - Handles rule range spanning all buckets

func TestFormatSwissPrice(t *testing.T)
// - Integers < 1000: "500" (no separator)
// - Integers >= 1000: "1%27000", "100%27000" (URL-encoded apostrophe)
// - Decimals: "99.99", "599.99"

func TestResolvePricePlaceholders(t *testing.T)
// - {min_price} and {max_price} replaced in URL
// - {price_buckets} replaced with formatted bucket string
// - {base64_start}...{base64_end} encoded
// - No rule for category → URL unchanged
```

### `dedup_test.go`

```go
func TestDeduplicateDeals(t *testing.T)
// - Same product from 2 shops → keeps cheapest
// - Different products → keeps both
// - Empty input → empty output
// - Case-insensitive matching

func TestMarkCheapest(t *testing.T)
// - Marks the cheapest across shops for each product name
// - Single product → marked
// - Non-deal products still participate in cheapest marking
```

### `cache_test.go`

```go
func TestResponseCache_PutGet(t *testing.T)
// - Put then get returns same data
// - Get before put returns not found
// - Expired entry returns not found (use short TTL + sleep)

func TestResponseCache_NilSafe(t *testing.T)
// - nil cache.get returns (nil, false)
// - nil cache.put is a no-op
```

### `enrich_test.go`

```go
func TestParsePriceResponse_Array(t *testing.T)
// - Conrad-style: productsPath + idPath + pricePath extracts prices
// - Missing price → product skipped

func TestParsePriceResponse_Map(t *testing.T)
// - Alltron-style: map keyed by SKU

func TestMergePrices(t *testing.T)
// - Products with matching IDs get prices merged
// - Products without matches are dropped
// - Title and image enrichment when paths configured

func TestBuildPriceRequestBody(t *testing.T)
// - {ids} placeholder replaced with comma-joined IDs
// - {articles} placeholder replaced with JSON array
```

## Dependency Strategy

**In-process.** All functions under test are pure transformations or use file I/O (cache) that can be tested with `t.TempDir()`. No mocks or adapters needed.

## Testing Strategy

- **New tests**: ~15-20 test functions covering the 5 untested files
- **Old tests to keep**: All existing tests remain — these are purely additive
- **Test fixtures**: Create `testdata/` files for price API response fixtures (Conrad array, Alltron map)
- **Priority order**: `prices_test.go` (most recently changed, highest regression risk) → `dedup_test.go` (notification correctness) → `enrich_test.go` (complex parsing) → `cache_test.go` (TTL edge cases)

## Implementation Recommendations

- Each test file should test its corresponding source file in isolation
- Use table-driven tests for `formatSwissPrice` and `buildPriceBuckets` — many small cases
- Cache tests should use `t.TempDir()` and short TTLs to avoid flaky time-based tests
- Enrich tests should use inline JSON strings or small fixture files, not full production API responses
- This is a prerequisite for the other refactors (#001, #003) — having test coverage first makes structural changes safer
