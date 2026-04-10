package cleaners_test

import (
	"testing"

	"github.com/trancee/DealScout/internal/parser/cleaners"
)

func TestNormalizeName(t *testing.T) {
	tests := []struct {
		category, input, want string
	}{
		{"smartphones", "Apple iPhone SE 2020", "Apple iPhone SE (2020)"},
		{"smartphones", "Apple iPhone SE3", "Apple iPhone SE (2022)"},
		{"smartphones", "APPLE IPHONE 8", "Apple iPhone 8"},
		{"smartphones", "SAMSUNG Galaxy A16", "Samsung Galaxy A16"},
		{"smartphones", "XIAOMI Redmi Note 14 Pro", "Xiaomi Redmi Note 14 Pro"},
		{"smartphones", "SAMSUNG Galaxy A16", "Samsung Galaxy A16"},
		{"smartphones", "Apple iPhone 16 Pro", "Apple iPhone 16 Pro"},
		{"smartphones", "NOKIA 225", "Nokia 225"},
		{"smartphones", "ONEPLUS Nord CE4 Lite", "OnePlus Nord CE4 Lite"},
		{"smartphones", "ONE PLUS Nord", "One Plus Nord"},
		{"smartphones", "Google Pixel 9a", "Google Pixel 9a"},
		{"smartphones", "HMD Arc", "HMD Arc"},
		{"smartphones", "ZTE Blade V70", "ZTE Blade V70"},
		{"smartphones", "ZTE Blade V70 Vita stone gray", "ZTE Blade V70 Vita"},
		{"smartphones", "ZTE Blade V70 stardust gray", "ZTE Blade V70"},
		{"smartphones", "ZTE BLADE A35E SILVERY GRAY 64 GB Silvery Gray Dual SIM", "ZTE Blade A35e"},
		{"smartphones", "Samsung Galaxy A17 A176", "Samsung Galaxy A17"},
		// {"Samsung Galaxy A30s Dual-SIM", "Samsung Galaxy A30s"},
		{"smartphones", "Samsung Galaxy XCover 7 EE", "Samsung Galaxy XCover 7 EE"},
		// {"Samsung SM-A175FZKBEUE", "Samsung Galaxy A17"},
		// {"Samsung Galaxy A52s 128 Black Refurbished C+", "Samsung Galaxy A52s"},
		{"smartphones", "  extra   spaces  ", "extra spaces"},
		{"smartphones", "SONIM TECHNOLOGIES XP100", "Sonim Technologies XP100"},
		{"smartphones", "CROSSCALL Core-S5", "Crosscall Core-S5"},

		// Fairphone
		{"smartphones", "The Fairphone (Gen 6)", "Fairphone 6"},

		// HMD
		// {"smartphones", "HMD ARC DS 4/64 SHADOW BLACK", "HMD Arc"},

		// motorola
		{"smartphones", "motorola pb7e0037se", "motorola edge 60 fusion"},

		// realme
		// {"smartphones", "REALME NOTE 70T 4+128GB OBSIDAN BLACK", "realme Note 70T"},
		{"smartphones", "realme Note70T obsidian black", "realme Note 70T"},

		/// notebooks
		{"notebooks", "Acer Aspire 16 Ai Oled (A16-52M-78VR) U7 256V", "Acer Aspire 16 AI OLED (A16-52M)"},
		{"notebooks", "Acer Nitro V 16 AI (ANV16-42-R0EV) RTX 5060", "Acer Nitro V 16 AI (ANV16-42)"},
		{"notebooks", "Acer Predator Helios Neo 16S AI OLED (PHN16S-71-96UU)", "Acer Predator Helios Neo 16S AI OLED (PHN16S-71)"},
		{"notebooks", "Acer Swift Air 16 Oled (SFA16-61M-R4WH)", "Acer Swift Air 16 OLED (SFA16-61M)"},
		{"notebooks", "Acer TravelMate P2 (TMP215-55-G2-TCO-7047)", "Acer TravelMate P2 (TMP215-55)"},
		{"notebooks", "Acer TravelMate B311 (TMB311-34-TCO-C32T)", "Acer TravelMate B311 (TMB311-34)"},
		{"notebooks", "Apple 13 MacBook Neo (A18 Pro) 5C GPU", "Apple MacBook Neo (13\", A18 Pro)"},
		{"notebooks", "Apple MacBook Air 13 2025 M4 10C GPU /", "Apple MacBook Air (13\", M4, 2025)"},
		{"notebooks", "Apple MacBook Air 15\" 2025 M4 10C Gpu /", "Apple MacBook Air (15\", M4, 2025)"},
		{"notebooks", "Apple MacBook Pro 14 M5 2025 10C CPU/10C GPU", "Apple MacBook Pro (14\", M5, 2025)"},
		{"notebooks", "Asus ExpertBook B3 (B3605CCA-MB0262X)", "ASUS ExpertBook B3 (B3605CCA)"},
		{"notebooks", "Asus TUF Gaming A18 FA808UP-S8072W", "ASUS TUF Gaming A18 (FA808UP)"},
		{"notebooks", "Asus VivoBook S 15 Oled (M5506WA-MA012W)", "ASUS Vivobook S 15 OLED (M5506WA)"},
		{"notebooks", "Asus Vivobook 16 Flip TP3607AA-SI009W", "ASUS Vivobook 16 Flip (TP3607AA)"},
		{"notebooks", "ASUS Vivobook S16 (M3607GA-SH002W)", "ASUS Vivobook S16 (M3607GA)"},
		{"notebooks", "Dell Pro 14 Plus (U5 235U", "Dell Pro 14 Plus (U5 235U)"},
		{"notebooks", "Dell Pro 16 Plus PB16250", "Dell Pro 16 Plus PB16250"},
		{"notebooks", "Hp Ai 15-fd2708nz", "HP AI 15-fd2708nz"},
		{"notebooks", "HP OMEN Transcend 16-u0700nz", "HP Omen Transcend 16-u0700nz"},
		{"notebooks", "Hp OmniBook 5 16-ba1728nz", "HP OmniBook 5 16-ba1728nz"},
		{"notebooks", "HP Omnibook X Flip 14 Next Gen AI PC BN4J4EA", "HP OmniBook X Flip 14"},
		{"notebooks", "HP ProBook 4 G1i 16 AI PC D0DL5ES", "HP ProBook 4 G1i 16"},
		{"notebooks", "Lenovo Chrome 14M9610", "Lenovo Chrome 14M9610"},
		{"notebooks", "Lenovo IdeaPad Slim 5 16AHP10 (Amd)", "Lenovo IdeaPad Slim 5 16AHP10 (AMD)"},
		{"notebooks", "Lenovo ThinkBook 16 G8 IAL", "Lenovo ThinkBook 16 G8"},
		{"notebooks", "Lenovo ThinkPad E16 Gen 3 (Intel)", "Lenovo ThinkPad E16 Gen 3 (Intel)"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := cleaners.NormalizeName(tt.input, tt.category)
			if got != tt.want {
				t.Errorf("NormalizeName(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}
