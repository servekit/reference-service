package data

import "testing"

func TestRegionGroupsIntegrity(t *testing.T) {
	var tops int
	for code, g := range RegionGroups {
		if g.Code != code {
			t.Errorf("key %q != code %q", code, g.Code)
		}
		if g.ParentCode == "" {
			tops++
		} else if _, ok := RegionGroups[g.ParentCode]; !ok {
			t.Errorf("%s: parent %s not a group", code, g.ParentCode)
		}
		if code == "001" {
			t.Errorf("world root 001 must not be served")
		}
	}
	if tops < 4 || tops > 8 {
		t.Fatalf("top-level groups %d outside [4,8]", tops)
	}
	for _, l := range Locales {
		if len(RegionGroupNames[l]) != len(RegionGroups) {
			t.Fatalf("locale %s: %d names for %d groups", l, len(RegionGroupNames[l]), len(RegionGroups))
		}
		if len(RegionGroupOrder[l]) != len(RegionGroups) {
			t.Fatalf("locale %s: order is not a permutation", l)
		}
	}
}

func TestRegionGroupsGoldenRows(t *testing.T) {
	asia := RegionGroups["142"]
	if asia.ParentCode != "" {
		t.Errorf("142 (Asia) parent = %q, want top level", asia.ParentCode)
	}
	if RegionGroupNames["zh-Hans"]["142"] != "亚洲" {
		t.Errorf("zh-Hans 142 = %q, want 亚洲", RegionGroupNames["zh-Hans"]["142"])
	}
	if RegionGroupNames["en"]["142"] != "Asia" {
		t.Errorf("en 142 = %q, want Asia", RegionGroupNames["en"]["142"])
	}
	ea := RegionGroups["030"] // Eastern Asia
	if ea.ParentCode != "142" || !contains(ea.RegionCodes, "CN") {
		t.Errorf("030 (Eastern Asia) wrong: %+v", ea)
	}
}
