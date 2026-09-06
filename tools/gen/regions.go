// regions.go — builds the region-group domain: the UN M49 hierarchy from
// CLDR territoryContainment (world root 001 excluded; top level =
// continents), parents resolved deterministically (lexicographically first
// containing group), and names from each locale's territories table.
package main

import (
	"fmt"
	"sort"
	"strings"
)

type regionRow struct {
	Code, ParentCode string
	CountryCodes     []string
}

func buildRegionGroups(cacheDir string) error {
	var tc territoryContainmentFile
	if err := getJSON(urlTerritoryContainment, cacheDir, &tc); err != nil {
		return err
	}
	containmentRaw := tc.Supplemental.TerritoryContainment

	// CLDR flattens per-member status into pseudo keys like
	// "001-status-grouping" (members EU/EZ/UN) or "017-status-deprecated"
	// (historical ZR). Split each key into base group + status; deprecated
	// memberships are dropped, others fold back into the base group.
	type group struct {
		contains map[string]bool
	}
	groupsMap := make(map[string]*group)
	addGroup := func(code string) *group {
		g, ok := groupsMap[code]
		if !ok {
			g = &group{contains: map[string]bool{}}
			groupsMap[code] = g
		}
		return g
	}
	for key, e := range containmentRaw {
		base, status := key, ""
		if i := strings.Index(key, "-status-"); i >= 0 {
			base, status = key[:i], key[i+len("-status-"):]
		}
		if status == "deprecated" {
			continue
		}
		g := addGroup(base)
		for _, m := range e.Contains {
			g.contains[m] = true
		}
	}

	// Served set: numeric M49 groups only, world root excluded. Letter-coded
	// CLDR containers (EU/EZ/UN/QO) are neither groups nor countries.
	isGroup := func(code string) bool {
		if code == "001" {
			return false
		}
		for _, r := range code {
			if r < '0' || r > '9' {
				return false
			}
		}
		return len(code) == 3
	}
	// Country members must exist in the country domain — same source.
	seedSet := make(map[string]bool, len(derivedDials()))
	for code := range derivedDials() {
		seedSet[code] = true
	}

	var groups []string
	for code := range groupsMap {
		if isGroup(code) {
			groups = append(groups, code)
		}
	}
	sort.Strings(groups)

	// Parent: a direct member of 001 is top level (""); otherwise the
	// lexicographically first other group that contains it.
	parent := make(map[string]string, len(groups))
	for _, g := range groups {
		if groupsMap["001"].contains[g] {
			parent[g] = ""
			continue
		}
		for _, container := range groups {
			if container != g && groupsMap[container].contains[g] {
				parent[g] = container
				break
			}
		}
	}

	entities := make(map[string]regionRow, len(groups))
	for _, g := range groups {
		var countries []string
		for _, m := range sortedKeysBool(groupsMap[g].contains) {
			if len(m) == 2 && seedSet[m] { // alpha-2 leaf members only
				countries = append(countries, m)
			}
		}
		entities[g] = regionRow{Code: g, ParentCode: parent[g], CountryCodes: countries}
	}

	// Names come from the same territories tables as countries; missing
	// entries fall back to English, then to the numeric code.
	names := make(map[string]map[string]string, len(locales))
	for _, l := range locales {
		var tf territoriesFile
		if err := getJSON(urlTerritories(l), cacheDir, &tf); err != nil {
			return fmt.Errorf("territories %s: %w", l, err)
		}
		tt := sole(tf.Main).LocaleDisplayNames.Territories
		names[l] = make(map[string]string, len(entities))
		for g := range entities {
			names[l][g] = tt[g]
		}
	}
	for _, l := range locales {
		for g, n := range names[l] {
			if n == "" {
				names[l][g] = firstNonEmpty(names["en"][g], g)
			}
		}
	}

	// Order: depth-first from top level (continents), siblings in the
	// locale's name order.
	orders := make(map[string][]string, len(locales))
	for _, l := range locales {
		orders[l] = regionOrder(entities, names[l])
	}

	w := newWriter("CLDR " + cldrVersion + " territoryContainment (UN M49 hierarchy)")
	w.line("// RegionGroups is the UN M49 group table keyed by numeric code;")
	w.line("// top level (ParentCode \"\") is a continent; the 001 world root is not served.")
	w.line("var RegionGroups = map[string]RegionGroup{")
	for _, g := range groups {
		e := entities[g]
		w.line("\t%q: {Code: %q, ParentCode: %q, CountryCodes: []string{%s}},",
			g, e.Code, e.ParentCode, quoteJoin(e.CountryCodes))
	}
	w.line("}")
	w.line("")
	w.emitNames("RegionGroupNames", names)
	w.emitOrder("RegionGroupOrder", orders)
	return w.save(outputPath("region_groups_data.go"))
}

// regionOrder walks the hierarchy depth-first; siblings sorted by their
// display name (plain string order keeps sibling sequences stable and is
// close enough for group labels; the primary sort requirement is structure).
func regionOrder(entities map[string]regionRow, names map[string]string) []string {
	children := make(map[string][]string)
	var tops []string
	for code, e := range entities {
		if e.ParentCode == "" {
			tops = append(tops, code)
		} else {
			children[e.ParentCode] = append(children[e.ParentCode], code)
		}
	}
	byName := func(codes []string) {
		sort.Slice(codes, func(i, j int) bool {
			if names[codes[i]] != names[codes[j]] {
				return names[codes[i]] < names[codes[j]]
			}
			return codes[i] < codes[j]
		})
	}
	byName(tops)
	var out []string
	var walk func([]string)
	walk = func(codes []string) {
		for _, c := range codes {
			out = append(out, c)
			kids := children[c]
			byName(kids)
			walk(kids)
		}
	}
	walk(tops)
	return out
}

// sortedKeysBool returns map keys in lexicographic order.
func sortedKeysBool(m map[string]bool) []string {
	ks := make([]string, 0, len(m))
	for k := range m {
		ks = append(ks, k)
	}
	sort.Strings(ks)
	return ks
}
