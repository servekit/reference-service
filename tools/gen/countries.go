// countries.go — builds the country domain: entity table from the seed
// (inclusion list + dial codes + zh/en names kept verbatim for migration
// parity), alpha-3 from the ISO 3166 dataset, names for the other locales
// from CLDR territories.
package main

import (
	"fmt"
)

// countryRow mirrors data.Country (the generator is a standalone main
// package and cannot import internal/data).
type countryRow struct {
	Code, Alpha3, DialCode, FlagEmoji string
}

func buildCountries(cacheDir string) error {
	var iso []isoRow
	if err := getJSON(urlISO3166, cacheDir, &iso); err != nil {
		return err
	}
	alpha3 := make(map[string]string, len(iso))
	for _, r := range iso {
		alpha3[r.Alpha2] = r.Alpha3
	}

	// CLDR territory names per locale (zh-Hans/en defer to the seed table).
	cldrNames := make(map[string]map[string]string, len(locales))
	for _, l := range locales {
		if l == "zh-Hans" || l == "en" {
			continue
		}
		var tf territoriesFile
		if err := getJSON(urlTerritories(l), cacheDir, &tf); err != nil {
			return fmt.Errorf("territories %s: %w", l, err)
		}
		cldrNames[l] = sole(tf.Main).LocaleDisplayNames.Territories
	}

	entities := make(map[string]countryRow, len(seedCountries))
	names := make(map[string]map[string]string, len(locales))
	for _, l := range locales {
		names[l] = make(map[string]string, len(seedCountries))
	}
	missing := map[string][]string{}

	for _, s := range seedCountries {
		entities[s.Code] = countryRow{
			Code:      s.Code,
			Alpha3:    alpha3[s.Code],
			DialCode:  s.Dial,
			FlagEmoji: flagEmoji(s.Code),
		}
		for _, l := range locales {
			var n string
			switch l {
			case "zh-Hans":
				n = s.NameZh
			case "en":
				n = s.NameEn
			default:
				n = cldrNames[l][s.Code]
				if n == "" {
					n = cldrNames[l][s.Code+"-alt-short"]
					// CLDR ships some territories with -alt- variants only.
					if n == "" {
						missing[l] = append(missing[l], s.Code)
						continue
					}
				}
			}
			names[l][s.Code] = n
		}
	}

	for l, codes := range missing {
		if len(codes) > 20 {
			return fmt.Errorf("locale %s: %d countries missing names (first: %v) — widen fallback", l, len(codes), codes[:3])
		}
		// A handful of exceptionally-reserved codes lack CLDR entries in
		// some locales; fall back to the English name rather than dropping
		// the row (List must stay complete across locales).
		for _, code := range codes {
			names[l][code] = seedName(code).NameEn
		}
	}

	orders := make(map[string][]string, len(locales))
	for _, l := range locales {
		orders[l] = collatedOrder(l, names[l])
	}

	w := newWriter("CLDR " + cldrVersion + " territories + ISO 3166 alpha-3 + seed dial codes (message-service parity)")
	w.line("// Countries is the entity table keyed by alpha-2.")
	w.line("var Countries = map[string]Country{")
	for _, code := range sortedKeys(entities) {
		c := entities[code]
		w.line("\t%q: {Code: %q, Alpha3: %q, DialCode: %q, FlagEmoji: %q},",
			code, c.Code, c.Alpha3, c.DialCode, c.FlagEmoji)
	}
	w.line("}")
	w.line("")
	w.emitNames("CountryNames", names)
	w.emitOrder("CountryOrder", orders)
	return w.save(outputPath("countries_data.go"))
}

func seedName(code string) seedCountry {
	for _, s := range seedCountries {
		if s.Code == code {
			return s
		}
	}
	return seedCountry{}
}

// flagEmoji maps alpha-2 to the Unicode regional-indicator pair.
func flagEmoji(code string) string {
	if len(code) != 2 {
		return ""
	}
	r := func(b byte) rune { return rune(0x1F1E6) + rune(b-'A') }
	return string(r(code[0])) + string(r(code[1]))
}
