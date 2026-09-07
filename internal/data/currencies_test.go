package data

import (
	"regexp"
	"testing"
)

func TestCurrenciesIntegrity(t *testing.T) {
	if n := len(Currencies); n < 150 || n > 200 {
		t.Fatalf("currency count %d outside [150,200]", n)
	}
	codeRe := regexp.MustCompile(`^[A-Z]{3}$`)
	for code, c := range Currencies {
		if c.Code != code || !codeRe.MatchString(code) {
			t.Errorf("bad code %q", code)
		}
		if c.MinorUnits < 0 || c.MinorUnits > 4 {
			t.Errorf("%s: minor_units %d outside [0,4]", code, c.MinorUnits)
		}
		if c.Symbol == "" {
			t.Errorf("%s: empty symbol", code)
		}
	}
	for _, l := range Locales {
		if len(CurrencyNames[l]) != len(Currencies) {
			t.Fatalf("locale %s: %d names for %d currencies", l, len(CurrencyNames[l]), len(Currencies))
		}
		if len(CurrencyOrder[l]) != len(Currencies) {
			t.Fatalf("locale %s: order is not a permutation", l)
		}
	}
}

func TestCurrenciesGoldenRows(t *testing.T) {
	if c := Currencies["CNY"]; c.Symbol != "¥" || c.MinorUnits != 2 || !contains(c.CountryCodes, "CN") {
		t.Errorf("CNY row wrong: %+v", c)
	}
	// Narrow symbols — one glyph, no disambiguation letters (the code column
	// disambiguates ¥/$ where needed).
	if c := Currencies["AUD"]; c.Symbol != "$" {
		t.Errorf("AUD symbol = %q, want $", c.Symbol)
	}
	if c := Currencies["USD"]; c.Symbol != "$" {
		t.Errorf("USD symbol = %q, want $", c.Symbol)
	}
	if c := Currencies["EUR"]; c.Symbol != "€" {
		t.Errorf("EUR symbol = %q, want €", c.Symbol)
	}
	if c := Currencies["JPY"]; c.MinorUnits != 0 {
		t.Errorf("JPY minor_units = %d, want 0", c.MinorUnits)
	}
	if c := Currencies["BHD"]; c.MinorUnits != 3 {
		t.Errorf("BHD minor_units = %d, want 3", c.MinorUnits)
	}
	if CurrencyNames["zh-Hans"]["USD"] != "美元" {
		t.Errorf("zh-Hans USD = %q", CurrencyNames["zh-Hans"]["USD"])
	}
	// Issuer flags: leading alpha-2 (incl. the exceptionally-reserved EU),
	// first member for X-prefixed multi-country codes, empty for XDR.
	if Currencies["CNY"].FlagEmoji != "\U0001F1E8\U0001F1F3" {
		t.Errorf("CNY flag = %q", Currencies["CNY"].FlagEmoji)
	}
	if Currencies["EUR"].FlagEmoji != "\U0001F1EA\U0001F1FA" {
		t.Errorf("EUR flag = %q", Currencies["EUR"].FlagEmoji)
	}
	if Currencies["XOF"].FlagEmoji == "" || Currencies["XDR"].FlagEmoji != "" {
		t.Errorf("XOF/XDR flags = %q/%q", Currencies["XOF"].FlagEmoji, Currencies["XDR"].FlagEmoji)
	}
	for code, c := range Currencies {
		if c.FlagEmoji != "" {
			r := []rune(c.FlagEmoji)
			if len(r) != 2 || r[0] < 0x1F1E6 || r[0] > 0x1F1FF || r[1] < 0x1F1E6 || r[1] > 0x1F1FF {
				t.Errorf("%s: bad flag %q", code, c.FlagEmoji)
			}
		}
	}
}
