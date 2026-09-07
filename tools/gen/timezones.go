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
	// zone1970.tab sorts by country code and puts each country's most
	// populous zone first — so the FIRST row mentioning a country is its
	// primary zone (the rule CLDR itself uses; primaryZones.json lists the
	// exceptions, overlaid below).
	primary := make(map[string]string)
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
		for _, cc := range strings.Split(fields[0], ",") {
			if _, ok := primary[cc]; !ok {
				primary[cc] = id
			}
		}
		if _, dup := entities[id]; dup {
			continue
		}
		codes := strings.Split(fields[0], ",")
		entities[id] = tzRow{ID: id, CountryCodes: codes}
	}

	var pz primaryZonesFile
	if err := getJSON(urlPrimaryZones, cacheDir, &pz); err != nil {
		return err
	}
	for cc, zone := range pz.Supplemental.PrimaryZones {
		if _, ok := entities[zone]; ok {
			primary[cc] = zone // CLDR exceptions are authoritative
		}
	}

	// Neither zone1970.tab nor zone.tab orders a country's zones by
	// population (within-country order is ~alphabetical: AU starts at
	// Lord_Howe, RU at Kaliningrad), and CLDR's primaryZones lists only 11
	// exceptions — so multi-zone countries outside that list get this
	// curated map, matching each country's most-populous-zone convention.
	// Single-zone countries are unambiguous (first-appearance above).
	for cc, zone := range curatedPrimaryZones {
		if _, ok := entities[zone]; !ok {
			return fmt.Errorf("curated primary zone %s for %s is not canonical", zone, cc)
		}
		primary[cc] = zone
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
			if city := exemplarCity(zones, id); city != "" {
				names[l][id] = city
				continue
			}
			// No exemplar city (Etc/*, golden zones) — display the last
			// path segment with underscores as spaces.
			last := id[strings.LastIndex(id, "/")+1:]
			names[l][id] = strings.ReplaceAll(last, "_", " ")
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
	w.line("// PrimaryZones maps alpha-2 -> the country's primary IANA zone")
	w.line("// (zone1970 first-row rule, CLDR primaryZones exceptions overlaid).")
	w.line("var PrimaryZones = map[string]string{")
	for _, cc := range sortedKeys(primary) {
		w.line("\t%q: %q,", cc, primary[cc])
	}
	w.line("}")
	w.line("")
	return w.save(outputPath("timezones_data.go"))
}

// curatedPrimaryZones covers every multi-zone country that CLDR's 11-entry
// exception list misses. Values follow the standard "most populous /
// capital-adjacent zone" convention (Wikipedia "Time in X" consensus).
var curatedPrimaryZones = map[string]string{
	"AR": "America/Argentina/Buenos_Aires",
	"AU": "Australia/Sydney",
	"BR": "America/Sao_Paulo",
	"CA": "America/Toronto",
	"CD": "Africa/Lagos", // Kinshasa (capital) merged as a link of the WAT group
	"CY": "Asia/Nicosia",
	"FM": "Pacific/Port_Moresby", // Chuuk (most populous); Pohnpei is a link to Guadalcanal
	"GL": "America/Nuuk",
	"ID": "Asia/Jakarta",
	"KI": "Pacific/Tarawa",
	"KZ": "Asia/Almaty",
	"MN": "Asia/Ulaanbaatar",
	"MX": "America/Mexico_City",
	"PG": "Pacific/Port_Moresby",
	"PS": "Asia/Gaza",
	"RU": "Europe/Moscow",
	"TF": "Indian/Maldives", // Kerguelen merged as a link (both UTC+5, no DST)
	"UM": "Pacific/Tarawa",  // Wake merged as a link (UTC+12 group)
	"US": "America/New_York",
	"VN": "Asia/Ho_Chi_Minh",
}
