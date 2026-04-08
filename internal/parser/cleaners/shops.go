package cleaners

import (
	"regexp"
	"strings"
)

var galaxusParenRe = regexp.MustCompile(`\s*\(.*\)\s*$`)

func cleanGalaxus(name string) string {
	// Galaxus format: "Brand Model (Storage, Color, Screen, SIM, Camera, Network)"
	// Extract key specs from parentheses before removing them.
	match := galaxusParenRe.FindString(name)
	base := galaxusParenRe.ReplaceAllString(name, "")

	if match == "" {
		return strings.TrimSpace(base)
	}

	// Parse parenthesized specs — keep storage and color (first two fields).
	inner := strings.Trim(match, " ()")
	parts := strings.Split(inner, ", ")

	var kept []string
	for i, part := range parts {
		if i >= 2 {
			break
		}
		kept = append(kept, strings.TrimSpace(part))
	}

	result := base
	if len(kept) > 0 {
		result += " " + strings.Join(kept, " ")
	}

	return strings.TrimSpace(result)
}

var amazonSuffixRe = regexp.MustCompile(`,\s*(Android|Smartphone|Handy|Mobiltelefon|Mobile Phone).*$`)

func cleanAmazon(name string) string {
	// Amazon format: "Brand Model, Android Smartphone, specs..."
	// Strip everything from the first category-indicator comma onward.
	name = amazonSuffixRe.ReplaceAllString(name, "")
	return strings.TrimSpace(name)
}
