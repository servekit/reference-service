package data

import "testing"

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
