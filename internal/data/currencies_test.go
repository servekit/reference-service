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
	if c := Currencies["CNY"]; c.Symbol != "CN¥" || c.MinorUnits != 2 || !contains(c.CountryCodes, "CN") {
		t.Errorf("CNY row wrong: %+v", c)
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
}
