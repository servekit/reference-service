// timezones.go — builds the timezone domain: canonical zones and country
// membership from IANA zone1970.tab, backward-link aliases from the tzdb
// backward file, and display names from CLDR exemplar cities (fallback:
// the zone id's last path segment).
package main

import (
	"fmt"
	"strings"
)

type tzRow struct {
	ID           string
	Aliases      []string
	CountryCodes []string
}

func buildTimezones(cacheDir string) error {
	tab, err := fetch(urlZoneTab, cacheDir)
	if err != nil {
		return err
	}
	back, err := fetch(urlTzBack, cacheDir)
	if err != nil {
		return err
	}

	entities := make(map[string]tzRow)
	for _, line := range strings.Split(string(tab), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		// Format: country-codes(,joined)  coordinates  TZ-id  comments
		fields := strings.Fields(line)
		if len(fields) < 3 {
			continue
		}
		id := fields[2]
		if _, dup := entities[id]; dup {
			continue
		}
		codes := strings.Split(fields[0], ",")
		entities[id] = tzRow{ID: id, CountryCodes: codes}
	}

	// Backward links: "Link <target> <alias>". Targets may themselves be
	// links; resolve transitively to a canonical zone.
	aliases := make(map[string]string)
	for _, line := range strings.Split(string(back), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) != 3 || fields[0] != "Link" {
			continue
		}
		target, alias := fields[1], fields[2]
		for hop := 0; hop < 5; hop++ {
			next, isLink := aliases[target]
			if !isLink {
				break
			}
			target = next
		}
		aliases[alias] = target
	}
	for alias, target := range aliases {
		row, ok := entities[target]
		if !ok {
			continue // link to a zone outside zone1970.tab (not canonical)
		}
		row.Aliases = append(row.Aliases, alias)
		entities[target] = row
	}

	// Names: CLDR exemplarCity under zone/<region>/<city>, fallback to the
	// id's last segment with underscores as spaces.
	names := make(map[string]map[string]string, len(locales))
	for _, l := range locales {
		var tf timezoneNamesFile
		if err := getJSON(urlTimezoneNames(l), cacheDir, &tf); err != nil {
			return fmt.Errorf("timezoneNames %s: %w", l, err)
		}
		zones := sole(tf.Main).Dates.TimeZoneNames.Zone
		names[l] = make(map[string]string, len(entities))
		for id := range entities {
			parts := strings.SplitN(id, "/", 2)
			if len(parts) != 2 {
				continue
			}
			if entry, ok := zones[parts[0]][parts[1]]; ok && entry.ExemplarCity != "" {
				names[l][id] = entry.ExemplarCity
				continue
			}
			// No exemplar city (Etc/*, golden zones, third-level ids like
			// America/Argentina/Buenos_Aires) — display the last segment.
			names[l][id] = strings.ReplaceAll(parts[len(parts)-1], "_", " ")
		}
	}

	orders := make(map[string][]string, len(locales))
	for _, l := range locales {
		orders[l] = collatedOrder(l, names[l])
	}

	w := newWriter("IANA zone1970.tab + tz backward links + CLDR " + cldrVersion + " exemplar cities")
	w.line("// Timezones is the canonical zone table keyed by IANA id.")
	w.line("var Timezones = map[string]Timezone{")
	for _, id := range sortedKeys(entities) {
		e := entities[id]
		w.line("\t%q: {ID: %q, Aliases: []string{%s}, CountryCodes: []string{%s}},",
			id, e.ID, quoteJoin(e.Aliases), quoteJoin(e.CountryCodes))
	}
	w.line("}")
	w.line("")
	w.line("// TimezoneAliases maps tzdb backward ids (and US/* style legacy names)")
	w.line("// to canonical zone ids — the input-normalization table.")
	w.line("var TimezoneAliases = map[string]string{")
	for _, a := range sortedKeys(aliases) {
		if _, ok := entities[aliases[a]]; ok {
			w.line("\t%q: %q,", a, aliases[a])
		}
	}
	w.line("}")
	w.line("")
	w.emitNames("TimezoneNames", names)
	w.emitOrder("TimezoneOrder", orders)
	return w.save(outputPath("timezones_data.go"))
}
