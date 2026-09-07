package data

import (
	"strings"
	"testing"
)

func TestTimezonesIntegrity(t *testing.T) {
	if n := len(Timezones); n < 300 || n > 450 {
		t.Fatalf("timezone count %d outside [300,450]", n)
	}
	sh := Timezones["Asia/Shanghai"]
	if sh.ID != "Asia/Shanghai" {
		t.Fatalf("Asia/Shanghai missing: %+v", sh)
	}
	if !contains(sh.Aliases, "PRC") || !contains(sh.Aliases, "Asia/Chongqing") {
		t.Fatalf("Asia/Shanghai aliases wrong: %v", sh.Aliases)
	}
	if !contains(sh.CountryCodes, "CN") {
		t.Fatalf("Asia/Shanghai countries wrong: %v", sh.CountryCodes)
	}
	for alias, target := range TimezoneAliases {
		if _, ok := Timezones[target]; !ok {
			t.Errorf("alias %s -> non-canonical %s", alias, target)
		}
		if _, dup := Timezones[alias]; dup {
			t.Errorf("alias %s collides with a canonical id", alias)
		}
	}
	if TimezoneAliases["PRC"] != "Asia/Shanghai" {
		t.Errorf("PRC -> %q, want Asia/Shanghai", TimezoneAliases["PRC"])
	}
	for _, l := range Locales {
		if len(TimezoneNames[l]) != len(Timezones) {
			t.Fatalf("locale %s: %d names for %d timezones", l, len(TimezoneNames[l]), len(Timezones))
		}
		if len(TimezoneOrder[l]) != len(Timezones) {
			t.Fatalf("locale %s: order is not a permutation", l)
		}
	}
	if TimezoneNames["zh-Hans"]["Asia/Shanghai"] != "上海" {
		t.Errorf("zh-Hans Shanghai = %q", TimezoneNames["zh-Hans"]["Asia/Shanghai"])
	}
	if TimezoneNames["en"]["Asia/Shanghai"] != "Shanghai" {
		t.Errorf("en Shanghai = %q", TimezoneNames["en"]["Asia/Shanghai"])
	}
}

func contains(ss []string, want string) bool {
	for _, s := range ss {
		if s == want {
			return true
		}
	}
	return false
}

// TestTimezoneNamesFallbackChain: zh-Hans names must not degrade to raw
// English city names — zones without a translated exemplarCity fall back
// to the metazone name (e.g. America/Argentina/* -> 阿根廷时间). Only
// Etc/GMT-style ids may stay Latin.
func TestTimezoneNamesFallbackChain(t *testing.T) {
	zh := TimezoneNames["zh-Hans"]
	if got := zh["America/Argentina/Buenos_Aires"]; got != "阿根廷时间" {
		t.Errorf("Buenos_Aires zh = %q, want 阿根廷时间", got)
	}
	if got := zh["America/Argentina/Rio_Gallegos"]; got != "里奥加耶戈斯" {
		t.Errorf("Rio_Gallegos zh = %q (exemplarCity must win over metazone)", got)
	}
	if got := zh["Asia/Shanghai"]; got != "上海" {
		t.Errorf("Shanghai zh = %q", got)
	}
	latin := 0
	for id, name := range zh {
		if strings.HasPrefix(id, "Etc/GMT") {
			continue
		}
		if allLatin(name) {
			latin++
			if latin <= 5 {
				t.Logf("latin-only zh name remains: %s -> %q", id, name)
			}
		}
	}
	// Known CLDR-45 gap: America/Coyhaique and America/Brit_Columbia/Golden
	// are tz-2025 additions whose metazone registrations haven't landed in
	// CLDR yet — they heal on the next CLDR bump.
	if latin > 2 {
		t.Fatalf("%d zh-Hans timezone names are still latin-only (want <=2, the CLDR-45 gap)", latin)
	}
}

func allLatin(s string) bool {
	for _, r := range s {
		if r >= 0x2E80 {
			return false
		}
	}
	return true
}
