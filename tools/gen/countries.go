// countries.go — builds the country domain from ONE phone-metadata source:
// the region set and dial codes are derived from nyaruka/phonenumbers (the
// same embedded metadata ParsePhone runs on — bumping the dependency and
// regenerating keeps the directory and the parser in lockstep), alpha-3
// comes from the ISO 3166 dataset, and names for every locale (including
// zh-Hans and en) come from CLDR territories.
package main

import (
	"fmt"
	"strconv"

	"github.com/nyaruka/phonenumbers"
)

// countryRow mirrors data.Country (the generator is a standalone main
// package and cannot import internal/data).
type countryRow struct {
	Code, Alpha3, DialCode, FlagEmoji string
}

// buildCountries emits the country entity table plus per-locale names and
// collated order.
func buildCountries(cacheDir string) error {
	var iso []isoRow
	if err := getJSON(urlISO3166, cacheDir, &iso); err != nil {
		return err
	}
	alpha3 := make(map[string]string, len(iso))
	for _, r := range iso {
		alpha3[r.Alpha2] = r.Alpha3
	}

	// CLDR territory names per locale — every locale including zh-Hans/en.
	cldrNames := make(map[string]map[string]string, len(locales))
	for _, l := range locales {
		var tf territoriesFile
		if err := getJSON(urlTerritories(l), cacheDir, &tf); err != nil {
			return fmt.Errorf("territories %s: %w", l, err)
		}
		cldrNames[l] = sole(tf.Main).LocaleDisplayNames.Territories
	}

	// Region set + dial codes from the phonenumbers metadata.
	dials := derivedDials()

	entities := make(map[string]countryRow, len(dials))
	names := make(map[string]map[string]string, len(locales))
	for _, l := range locales {
		names[l] = make(map[string]string, len(dials))
	}
	missing := map[string][]string{}

	for code, dial := range dials {
		entities[code] = countryRow{
			Code:      code,
			Alpha3:    alpha3[code],
			DialCode:  dial,
			FlagEmoji: flagEmoji(code),
		}
		for _, l := range locales {
			n := cldrNames[l][code]
			if n == "" {
				n = cldrNames[l][code+"-alt-short"] // some territories only have -alt variants
			}
			if n == "" {
				missing[l] = append(missing[l], code)
				continue
			}
			names[l][code] = n
		}
	}

	for l, codes := range missing {
		if len(codes) > 20 {
			return fmt.Errorf("locale %s: %d countries missing names (first: %v) — widen fallback", l, len(codes), codes[:3])
		}
		// A handful of exceptionally-reserved codes lack CLDR entries in
		// some locales; fall back to English (then the code itself) rather
		// than dropping the row (List must stay complete across locales).
		for _, code := range codes {
			names[l][code] = firstNonEmpty(cldrNames["en"][code], code)
		}
	}

	orders := make(map[string][]string, len(locales))
	for _, l := range locales {
		orders[l] = collatedOrder(l, names[l])
	}

	w := newWriter("nyaruka/phonenumbers region metadata (dial codes) + ISO 3166 alpha-3 + CLDR " + cldrVersion + " territories")
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

// derivedDials returns region -> "+<calling code>" from the same metadata
// ParsePhone runs on — the single phone-metadata source of this service.
func derivedDials() map[string]string {
	out := make(map[string]string)
	for region := range phonenumbers.GetSupportedRegions() {
		if len(region) != 2 {
			continue
		}
		cc := phonenumbers.GetCountryCodeForRegion(region)
		if cc <= 0 {
			continue
		}
		out[region] = "+" + strconv.Itoa(cc)
	}
	return out
}

// flagEmoji maps alpha-2 to the Unicode regional-indicator pair.
func flagEmoji(code string) string {
	if len(code) != 2 {
		return ""
	}
	r := func(b byte) rune { return rune(0x1F1E6) + rune(b-'A') }
	return string(r(code[0])) + string(r(code[1]))
}
