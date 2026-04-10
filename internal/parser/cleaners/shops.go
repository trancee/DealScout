package cleaners

import (
	"regexp"
	"strings"
)

var ackermannGuillemetsRe = regexp.MustCompile(`»(.*?)«`)
var ackermannSpecSuffixRe = regexp.MustCompile(`(?i)(,\s*)?\d+\s*GB|(,\s*)?\(?[2345]G\)?`)

func cleanAckermann(name string) string {
	// Ackermann format: "Brand Type »Model Specs« extra" or "Type »Model Specs« extra"
	// Extract the content between guillemets, preserving the brand prefix.
	if matches := ackermannGuillemetsRe.FindStringSubmatch(name); len(matches) > 1 {
		inner := matches[1]
		// Find everything before the guillemets — the brand is the first word if present.
		before := name[:strings.Index(name, "»")]
		before = strings.TrimRight(before, " ")
		// Remove type words (Smartphone, Handy, etc.) to isolate the brand.
		before = strings.NewReplacer("Smartphone", "", "Handy", "", "Mobiltelefon", "", "Chromebook", "", "Business-Notebook", "", "Gaming-Notebook", "", "Convertible Notebook", "").Replace(before)
		before = strings.TrimSpace(before)
		if before != "" && !strings.HasPrefix(strings.ToLower(inner), strings.ToLower(before)) {
			name = before + " " + inner
		} else {
			name = inner
		}
	}

	// notebooks
	name = strings.NewReplacer("Apple Notebook Air", "Apple MacBook Air", "Notebook", "").Replace(name)

	model := regexp.MustCompile(`(\(.*?\))`).FindString(name)

	// Strip storage/network suffixes (e.g., "128 GB", "5G", "LTE").
	if loc := ackermannSpecSuffixRe.FindStringIndex(name); loc != nil {
		name = name[:loc[0]]

		if !strings.Contains(name, "(") {
			name += " " + model
		}
	}

	return strings.TrimSpace(name)
}

var alltronSpecRe = regexp.MustCompile(`(\s*[-,]\s+)|(\b\d{1,3}\s*GB?\b)|Copilot|\s+CH$`)

func cleanAlltron(name string) string {
	// smartphones
	name = strings.NewReplacer("Enterprise Edition", "EE", "Fairphone Fairphone", "Fairphone", "EU-Ware", "").Replace(name)
	// notebooks
	name = strings.NewReplacer("Notebook", "").Replace(name)

	if loc := alltronSpecRe.FindStringSubmatchIndex(name); loc != nil {
		name = name[:loc[0]]
	}

	return strings.TrimSpace(name)
}

var amazonSuffixRe = regexp.MustCompile(`(?i)[,|]\s*(Android|Smartphone|Handy|Mobiltelefon|Mobile Phone|Telefon|Dual[ -]?SIM|Entsperrt|ohne Vertrag|Unlocked|Simlockfrei|Global Version)|without simlock|Dual SIM|(SIM )?Smartphone|Mobile Phone|\bEU\b|\b[45]G\b|\b(8|128|256)GB\b| - `)

func cleanAmazon(name string) string {
	// Strip everything from the first category/spec indicator onward.
	if loc := amazonSuffixRe.FindStringIndex(name); loc != nil {
		name = name[:loc[0]]
	}

	// Strip parenthesized specs like "(128GB, Black)".
	name = regexp.MustCompile(`\s*\([^)]*\)|Compatible with`).ReplaceAllString(name, "")

	return strings.TrimSpace(name)
}

var brackSpecRe = regexp.MustCompile(`(\s*[-,]\s+)|(\b\d{1,3}\s*GB?\b)|Copilot|\s+CH$`)

func cleanBrack(name string) string {
	// smartphones
	name = strings.NewReplacer("Enterprise Edition", "EE", "Fairphone Fairphone", "Fairphone", "EU-Ware", "", "EU-Version", "").Replace(name)
	// notebooks
	name = strings.NewReplacer("Notebook", "").Replace(name)

	if loc := brackSpecRe.FindStringSubmatchIndex(name); loc != nil {
		name = name[:loc[0]]
	}

	return strings.TrimSpace(name)
}

var conforamaSpecRe = regexp.MustCompile(`(?i)(\d+/)?(1|2|4|6|8|16|32|64|128|256|512)\s*[GT]B|\(?[2345]G\)?|R-5|Integrated|LTE|Dual([- ]SIM)?`)

func cleanConforama(name string) string {
	// Remove duplicated brand prefix (e.g. "ZTE ZTE Blade A35" → "ZTE Blade A35").
	words := strings.SplitN(name, " ", 3)
	if len(words) >= 2 && strings.EqualFold(words[0], words[1]) {
		name = words[0] + " " + strings.Join(words[2:], " ")
	}

	// name = strings.NewReplacer("''", "\"").Replace(name)
	name = regexp.MustCompile(`\s+Notebook \d+(\.\d+)?''`).ReplaceAllString(name, "")

	// Strip storage/network suffixes.
	if loc := conforamaSpecRe.FindStringIndex(name); loc != nil {
		name = name[:loc[0]]
	}

	return strings.TrimSpace(name)
}

var conradSpecRe = regexp.MustCompile(`\s*[-,]\s+|\W\+\s+|EU |\d+\s*[GM]B|\s*\d+G|\s+\(Version 20[12]\d\)|\s+\(Grade [A-Z]\)|\s+(((Senioren-|senior |Industrie |Outdoor )?Smartphone)|\s*CH$|Satellite|Ex-geschütztes Handy|Fusion( Holiday Edition)?|Refurbished|\(PRODUCT\) RED™|Weiß)`)

func cleanConrad(name string) string {
	name = strings.NewReplacer(
		"Enterprise Edition", "EE",
		"Renewd® ", "",
		"refurbished", "",
		"5G Smartphone", "",
		"Samsung XCover", "Samsung Galaxy XCover",
		"Edge20", "Edge 20",
		"Edge Neo 40", "Edge 40 Neo",
		"EU-Ware", "",
		"EU-Version", "",
		" (EU)", "",
	).Replace(name)

	// Remove duplicate brand prefix (e.g., "Nokia Nokia 105")
	parts := strings.SplitN(name, " ", 3)
	if len(parts) >= 2 && strings.EqualFold(parts[0], parts[1]) {
		name = parts[0] + " " + strings.Join(parts[2:], " ")
	}

	if loc := conradSpecRe.FindStringSubmatchIndex(name); loc != nil {
		name = name[:loc[0]]
	}

	return strings.TrimSpace(name)
}

var folettiSpecRe = regexp.MustCompile(`(?i)\s*[-,]+\s+|\s*\(?(\d+(\s*GB)?[+/])?\d+\s*GB\)?|\s*[45]G|(2|4|6|8|12)/(64|128|256?B?)(GB)?|\s+\(?20[12]\d\)?|\s*\d+([,.]\d+)?\s*(cm|inch|\")|\d{4,5}\s*mAh|\s+20[12]\d|\s+(Hybrid|Dual\W(SIM|Sim)|\s*CH( -|$)|inkl\.|LTE|NFC|smartphone)`)

func cleanFoletti(name string) string {
	name = strings.NewReplacer("Enterprise Edition", "EE", "Enterprise", "EE", "Renewd ", "", "SMARTPHONE ", "", "Smartphone ", "", "Smartfon ", "", "EU-Ware", "", "EU-Version", "", "Motorola Mobility", "").Replace(name)

	if loc := folettiSpecRe.FindStringSubmatchIndex(name); loc != nil {
		name = name[:loc[0]]
	}

	return strings.TrimSpace(name)
}

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

var interdiscountSpecRe = regexp.MustCompile(`\(\d+(\.\d+)?\s*[GM]B?|\(\d\.\d{1,2}"|\s+[2345]G| LTE`)

func cleanInterdiscount(name string) string {
	name = strings.NewReplacer("Enterprise Edition", "EE").Replace(name)

	if loc := interdiscountSpecRe.FindStringSubmatchIndex(name); loc != nil {
		name = name[:loc[0]]
	}

	// Remove duplicate brand prefix (e.g., "NOKIA Nokia 105")
	parts := strings.SplitN(name, " ", 3)
	if len(parts) >= 2 && strings.EqualFold(parts[0], parts[1]) {
		name = parts[0] + " " + strings.Join(parts[2:], " ")
	}

	return strings.TrimSpace(name)
}

var mediamarktSpecRe = regexp.MustCompile(` - |\d+\s*G[Bb]|\s+[2345]G|Dual-SIM|\(EU\)|\s+CH$`)

func cleanMediamarkt(name string) string {
	name = strings.NewReplacer("ONE PLUS", "ONEPLUS", "Enterprise Edition", "EE", "6941764469662", "Note 70T", "13024998", "X6B").Replace(name)

	if loc := mediamarktSpecRe.FindStringSubmatchIndex(name); loc != nil {
		name = name[:loc[0]]
	}

	return strings.TrimSpace(name)
}

var mobilezoneSpecRe = regexp.MustCompile(`\s+\(?\d+\s*GB?|\s+\(?\d+(\.\d+)?"| Dual Sim`)

func cleanMobilezone(name string) string {
	name = strings.NewReplacer(" Xcover5", " XCover 5", " 5G", "", " 128 ", " ").Replace(name)

	if loc := mobilezoneSpecRe.FindStringSubmatchIndex(name); loc != nil {
		name = name[:loc[0]]
	}

	return strings.TrimSpace(name)
}

var orderflowSpecRe = regexp.MustCompile(`\s+\(?(\d\+)?\d+\s*GB?|\s+\(?\d+(\.\d+)?"|\s+\(?[2345]G\)?| Dual SIM|, |\s*CH$`)

func cleanOrderflow(name string) string {
	name = strings.NewReplacer("Motorola Mobility ", "", "Enterprise Edition", "EE", " 4G ", " ", "EU-Ware", "", "EU-Version", "").Replace(name)

	if loc := orderflowSpecRe.FindStringSubmatchIndex(name); loc != nil {
		name = name[:loc[0]]
	}

	return strings.TrimSpace(name)
}

var postShopSpecRe = regexp.MustCompile(`(?i)(\d+[/+])?(1|2|4|6|8|16|32|64|128|256|512|1000)\s*[GT]B|(DS\s+)?(\d+[/+]\d+)|\(?[2345]G\)?|Bundle|Reenova|LTE|Dual([- ]SIM)?`)

func cleanPostShop(name string) string {
	// Strip parenthesized specs like "(128GB, Black)".
	name = regexp.MustCompile(`\s*\([^)]*\)`).ReplaceAllString(name, "")

	// Strip storage/network suffixes.
	if loc := postShopSpecRe.FindStringIndex(name); loc != nil {
		name = name[:loc[0]]
	}

	return strings.TrimSpace(name)
}

// var cashConvertersSpecRe = regexp.MustCompile(`(?i),|(\d+/)?(2|4|6|8|16|32|64|128|256|512)\s*(GB|Go|G|B)\b|\d'|\(?[345]G\)?|NFC|LTE|Dual([- ]SIM)?|NEUF|\+? Boîte`)

// func cleanCashConverters(name string) string {
// 	name = strings.NewReplacer("One +", "OnePlus", " - ", " ").Replace(name)
// 	name = regexp.MustCompile(`(?i)Portable|(Samsung )?Reconditionné|\(?(Blanc|Rouge)\)?|Téléphone(\s*:\s*)?|: `).ReplaceAllString(name, "")
// 	name = strings.TrimSpace(name)

// 	if loc := cashConvertersSpecRe.FindStringSubmatchIndex(name); loc != nil {
// 		name = name[:loc[0]]
// 	}

// 	name = strings.NewReplacer(
// 		"Samsung Samsung", "Samsung",
// 		"Samsung Note", "Samsung Galaxy Note",
// 		"Samsung XCOVER", "Samsung Galaxy XCover",
// 		"S20FE", "S20 FE",
// 	).Replace(name)

// 	return strings.TrimSpace(name)
// }

// var hopCashSpecRe = regexp.MustCompile(`(?i),|(\d+/)?(2|4|6|8|16|32|64|128|256|512)\s*(GB|Go|G|B)\b|\d'|\(?[345]G\)?|NFC|LTE|Dual([- ]SIM)?|NEUF|\+? Boîte`)

// func cleanHopCash(name string) string {
// 	name = strings.NewReplacer("One +", "OnePlus", " - ", " ").Replace(name)
// 	name = regexp.MustCompile(`(?i)Portable|(Samsung )?Reconditionné|\(?(Blanc|Rouge|Noir|Bleu|Vert|Gris|Rose)\)?|Téléphone(\s*:\s*)?|: |Smartphone |Galaaxy `).ReplaceAllString(name, "")
// 	name = strings.TrimSpace(name)

// 	if loc := hopCashSpecRe.FindStringSubmatchIndex(name); loc != nil {
// 		name = name[:loc[0]]
// 	}

// 	name = strings.NewReplacer(
// 		"Samsung Samsung", "Samsung",
// 		"Samsung Note", "Samsung Galaxy Note",
// 		"Samsung XCOVER", "Samsung Galaxy XCover",
// 	).Replace(name)

// 	return strings.TrimSpace(name)
// }
