package cleaners

import (
	"regexp"
	"strings"

	"github.com/trancee/DealScout/internal/config"
)

// NewFilter creates a FilterFunc from a config.Filter.
// Returns a function that returns true if the product name should be skipped.
func NewFilter(f config.Filter) FilterFunc {
	var exclusionRe *regexp.Regexp
	if f.ExclusionRegex != "" {
		exclusionRe = regexp.MustCompile(f.ExclusionRegex)
	}

	// Lowercase skip brands for case-insensitive matching.
	skipBrands := make([]string, len(f.SkipBrands))
	for i, b := range f.SkipBrands {
		skipBrands[i] = strings.ToLower(b)
	}

	return func(name string) bool {
		lower := strings.ToLower(name)

		for _, brand := range skipBrands {
			if strings.HasPrefix(lower, brand+" ") || strings.HasPrefix(lower, brand+"\t") || lower == brand {
				return true
			}
		}

		if exclusionRe != nil && exclusionRe.MatchString(name) {
			return true
		}

		return false
	}
}
