package data

import "testing"

// TestNoDanglingRegionCodes guards the directory's self-consistency: no
// emitted table may reference an alpha-2 code the Countries table does not
// serve. The generators filter against the dialing-region set (regions.go
// always did; languages/timezones/currencies joined after a review found
// AQ/TF/PN/GS/UM/EA/IC/... leaking in from upstream territories).
func TestNoDanglingRegionCodes(t *testing.T) {
	dangling := func(where string, codes ...string) {
		for _, cc := range codes {
			if _, ok := Countries[cc]; !ok {
				t.Errorf("%s references unknown country %q", where, cc)
			}
		}
	}
	for code, cur := range Currencies {
		dangling("currency "+code, cur.RegionCodes...)
	}
	for id, tz := range Timezones {
		dangling("timezone "+id, tz.RegionCodes...)
	}
	for cc := range PrimaryZones {
		dangling("PrimaryZones", cc)
	}
	for cc := range CountryLanguages {
		dangling("CountryLanguages", cc)
	}
	for g, rg := range RegionGroups {
		dangling("region group "+g, rg.RegionCodes...)
	}
}

// TestRegionGroupsAcyclic walks every parent chain to the root with a step
// bound; the bound failing means the hierarchy gained a cycle, which would
// hang the unbounded service-side walks.
func TestRegionGroupsAcyclic(t *testing.T) {
	for g := range RegionGroups {
		seen := map[string]bool{g: true}
		cur := RegionGroups[g].ParentCode
		for steps := 0; cur != ""; steps++ {
			if steps > len(RegionGroups) {
				t.Fatalf("region group %s: parent chain exceeds group count (cycle?)", g)
			}
			if seen[cur] {
				t.Fatalf("region group %s: cycle through %s", g, cur)
			}
			seen[cur] = true
			cur = RegionGroups[cur].ParentCode
		}
	}
}
