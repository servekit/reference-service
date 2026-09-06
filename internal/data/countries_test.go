package data

import (
	"regexp"
	"testing"
)

func TestCountriesIntegrity(t *testing.T) {
	if n := len(Countries); n < 240 || n > 260 {
		t.Fatalf("country count %d outside [240,260]", n)
	}
	alpha3 := regexp.MustCompile(`^[A-Z]{3}$`)
	dial := regexp.MustCompile(`^\+[0-9]+$`)
	noAlpha3 := 0
	for code, c := range Countries {
		if c.Code != code {
			t.Errorf("key %q != code %q", code, c.Code)
		}
		if c.Alpha3 != "" {
			if !alpha3.MatchString(c.Alpha3) {
				t.Errorf("%s: bad alpha3 %q", code, c.Alpha3)
			}
		} else {
			noAlpha3++ // exceptionally-reserved codes have no ISO alpha-3
		}
		if !dial.MatchString(c.DialCode) {
			t.Errorf("%s: bad dial %q", code, c.DialCode)
		}
		r := []rune(c.FlagEmoji)
		if len(r) != 2 || r[0] < 0x1F1E6 || r[0] > 0x1F1FF || r[1] < 0x1F1E6 || r[1] > 0x1F1FF {
			t.Errorf("%s: bad flag %q", code, c.FlagEmoji)
		}
	}
	if noAlpha3 > 10 {
		t.Fatalf("%d countries without alpha3 — too many", noAlpha3)
	}
	for _, l := range Locales {
		names := CountryNames[l]
		if len(names) != len(Countries) {
			t.Fatalf("locale %s: %d names for %d countries", l, len(names), len(Countries))
		}
		for c, n := range names {
			if n == "" {
				t.Errorf("locale %s: empty name for %s", l, c)
			}
		}
		if len(CountryOrder[l]) != len(Countries) {
			t.Fatalf("locale %s: order is not a permutation (%d vs %d)",
				l, len(CountryOrder[l]), len(Countries))
		}
		seen := make(map[string]bool, len(Countries))
		for _, c := range CountryOrder[l] {
			if seen[c] {
				t.Fatalf("locale %s: duplicate %s in order", l, c)
			}
			seen[c] = true
		}
	}
}

func TestCountriesGoldenRows(t *testing.T) {
	cn := Countries["CN"]
	if cn.Alpha3 != "CHN" || cn.DialCode != "+86" || cn.FlagEmoji != "\U0001F1E8\U0001F1F3" {
		t.Fatalf("CN row wrong: %+v", cn)
	}
	if CountryNames["zh-Hans"]["CN"] != "中国" {
		t.Errorf("zh-Hans CN = %q", CountryNames["zh-Hans"]["CN"])
	}
	if CountryNames["en"]["CN"] != "China" {
		t.Errorf("en CN = %q", CountryNames["en"]["CN"])
	}
	if CountryNames["ja"]["CN"] != "中国" {
		t.Errorf("ja CN = %q", CountryNames["ja"]["CN"])
	}
}
