# Refactor: Split productNormalizer into per-brand normalizer registry

## Problem

`productNormalizer` in `normalize.go` is a 350-line function containing 31 `if brand == "..."` blocks, each with brand-specific regex rules. The function is the single largest in the codebase (418 lines including helpers).

Problems:
- **Adding/fixing a brand** requires editing a massive function and visually scanning past 30 other brands
- **Fall-through mutations** — some brands (e.g. `galaxy` → `samsung`, `poco`/`redmi`/`mi` → `xiaomi`) mutate `brand` and rely on a later `if` block to run, creating implicit ordering dependencies
- **Category gating** — the entire function is wrapped in `switch category { case "smartphones": ... case "notebooks": ... }`, mixing category routing with brand logic
- **Test granularity** — `normalize_test.go` tests the whole pipeline; a Samsung regex bug shows up as a failure in a test case that doesn't mention Samsung in its name

## Proposed Interface

Replace the monolithic function with a registry of brand normalizers:

```go
// BrandNormalizer transforms a product name for a specific brand.
type BrandNormalizer func(name string, words []string) string

// brandNormalizers maps lowercase brand names to their normalizer functions.
var brandNormalizers = map[string]BrandNormalizer{
    "apple":    normalizeApple,
    "samsung":  normalizeSamsung,
    "xiaomi":   normalizeXiaomi,
    // ...
}

// Brand aliases that map to another brand's normalizer after name transformation.
var brandAliases = map[string]string{
    "iphone": "apple",   // prepend "Apple"
    "galaxy": "samsung", // prepend "Samsung"
    "poco":   "xiaomi",  // prepend "Xiaomi"
    "redmi":  "xiaomi",
    "mi":     "xiaomi",
    "moto":   "motorola",
}
```

Each normalizer lives in its own file (e.g., `normalize_samsung.go`, `normalize_apple.go`) or is grouped by complexity (small brands can share a file).

The dispatcher:
```go
func productNormalizer(name, category string) string {
    words := strings.Split(name, " ")
    brand := strings.ToLower(words[0])

    // Handle prefix words like "The"
    if brand == "the" && len(words) > 1 {
        name = strings.Join(words[1:], " ")
        words = strings.Split(name, " ")
        brand = strings.ToLower(words[0])
    }

    // Resolve aliases (prepend canonical brand)
    if target, ok := brandAliases[brand]; ok {
        name = canonicalBrandName(target) + " " + name
        words = strings.Split(name, " ")
        brand = target
    }

    // Dispatch to brand normalizer
    if fn, ok := brandNormalizers[brand]; ok {
        name = fn(name, words)
    }

    return strings.TrimSpace(name)
}
```

## Dependency Strategy

**In-process.** Pure string transformation with no I/O. Each brand normalizer is a stateless function operating on the product name string.

## Testing Strategy

- **New boundary tests**: Each brand normalizer gets its own test function (e.g., `TestNormalizeSamsung`, `TestNormalizeApple`) with brand-specific test tables. Failures immediately identify the broken brand.
- **Old tests to keep**: The existing `TestNormalizeName` table becomes an integration test verifying the full pipeline (dispatcher + normalizers + nameMapping). It should still pass unchanged.
- **Test environment**: None needed — pure functions.

## Implementation Recommendations

- The normalizer registry should own: brand detection, alias resolution, dispatching to the correct normalizer
- It should hide: the number of brands, per-brand regex complexity, fall-through mutation patterns
- It should expose: `productNormalizer(name, category string) string` — same signature as today
- Brand aliases should replace the current fall-through pattern (`if brand == "galaxy" { name = "Samsung " + name }` → alias entry)
- The `category` parameter should gate which normalizer registry is active (smartphones vs notebooks may need different brand rules)
- Consider pre-compiling regexes as package-level vars within each brand file instead of compiling inside the function body on every call
