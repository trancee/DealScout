package cleaners

import (
	"regexp"
	"strings"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

var nameMapping = map[string]string{
	"Ee": "EE", "Fe": "FE", "Gt": "GT", "Hd": "HD",
	"Htc": "HTC", "Iphone": "iPhone", "Lg": "LG",
	"Oneplus": "OnePlus", "Se": "SE", "Tcl": "TCL",
	"Zte": "ZTE", "Xcover": "XCover", "Xl": "XL",
	"Xr": "XR", "Xs": "XS", "Hmd": "HMD", "Nfc": "NFC",
	"Tecno": "TECNO", "Umidigi": "UMIDIGI",
}

var titleCaser = cases.Title(language.Und, cases.NoLower)
var multiSpaceRe = regexp.MustCompile(`\s{2,}`)

// NormalizeName converts an ALL CAPS or mixed-case product name to a
// consistent title-case form with brand-specific corrections.
func NormalizeName(name string) string {
	name = strings.TrimSpace(name)
	if name == "" {
		return name
	}

	// Collapse whitespace.
	name = multiSpaceRe.ReplaceAllString(name, " ")

	// If all uppercase or all lowercase, title-case it.
	upper := strings.ToUpper(name)
	lower := strings.ToLower(name)
	if name == upper || name == lower {
		name = titleCaser.String(lower)
	}

	// Apply brand/model corrections.
	for wrong, right := range nameMapping {
		if strings.Contains(name, wrong) {
			name = replaceWord(name, wrong, right)
		}
	}

	return strings.TrimSpace(name)
}

func replaceWord(s, old, new string) string {
	words := strings.Split(s, " ")
	for i, w := range words {
		if w == old {
			words[i] = new
		}
	}
	return strings.Join(words, " ")
}
