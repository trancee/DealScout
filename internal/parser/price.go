package parser

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

var (
	// Matches digits, dots, commas, apostrophes, and dash (for "119.–").
	priceDigitsRe = regexp.MustCompile(`[\d]['\d.,–-]*[\d–-]|[\d]+`)
)

// ParsePrice parses a European-format price string and returns a float64.
// Handles: "CHF 119.–", "€ 99,90", "1'299.00", "1.299,00", "1,299.00", etc.
func ParsePrice(s string) (float64, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, fmt.Errorf("empty price string")
	}

	match := priceDigitsRe.FindString(s)
	if match == "" {
		return 0, fmt.Errorf("no numeric value found in %q", s)
	}

	// Replace trailing dash/en-dash (e.g., "119.–" → "119.")
	match = strings.TrimRight(match, "–-")
	match = strings.TrimRight(match, ".")

	// Remove apostrophes (Swiss thousand separator: 1'299)
	match = strings.ReplaceAll(match, "'", "")

	// Determine if comma or dot is the decimal separator.
	lastDot := strings.LastIndex(match, ".")
	lastComma := strings.LastIndex(match, ",")

	switch {
	case lastDot == -1 && lastComma == -1:
		// Pure integer: "1299"
	case lastDot == -1 && lastComma != -1:
		// Comma is decimal: "99,90" or thousand sep: "1,299"
		// If exactly 3 digits after comma and no other commas → thousand separator
		afterComma := match[lastComma+1:]
		if len(afterComma) == 3 && strings.Count(match, ",") == 1 {
			match = strings.ReplaceAll(match, ",", "")
		} else {
			match = strings.ReplaceAll(match, ",", ".")
		}
	case lastDot != -1 && lastComma == -1:
		// Dot is decimal: "119.00" or thousand sep: "1.299"
		afterDot := match[lastDot+1:]
		if len(afterDot) == 3 && strings.Count(match, ".") == 1 {
			// Ambiguous: could be "1.299" (thousand) or "0.999"
			// Check if there are digits before the dot suggesting thousands
			beforeDot := match[:lastDot]
			if len(beforeDot) >= 1 && len(beforeDot) <= 3 {
				// Likely thousand separator for values like "1.299"
				// But "0.99" should be decimal — check if before is "0"
				val, _ := strconv.ParseFloat(match, 64)
				if val < 1 {
					// Keep as decimal
				} else {
					match = strings.ReplaceAll(match, ".", "")
				}
			}
		}
		// Otherwise dot is decimal separator — keep as is
	case lastComma > lastDot:
		// "1.299,00" → dot is thousand, comma is decimal
		match = strings.ReplaceAll(match, ".", "")
		match = strings.ReplaceAll(match, ",", ".")
	case lastDot > lastComma:
		// "1,299.00" → comma is thousand, dot is decimal
		match = strings.ReplaceAll(match, ",", "")
	}

	val, err := strconv.ParseFloat(match, 64)
	if err != nil {
		return 0, fmt.Errorf("parsing %q as float: %w", match, err)
	}
	return val, nil
}
