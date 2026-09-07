// languages.go — builds the language domain: the selectable BCP 47 tag set
// is the two-letter (ISO 639-1) subset that zh-Hans and en both name, plus
// the script-qualified tags we serve (zh-Hans, zh-Hant); names come from
// each locale's CLDR languages table.
package main

import (
	"fmt"
	"sort"
	"strconv"
)

func buildLanguages(cacheDir string) error {
	var zh, en languagesFile
	if err := getJSON(urlLanguages("zh-Hans"), cacheDir, &zh); err != nil {
		return err
	}
	if err := getJSON(urlLanguages("en"), cacheDir, &en); err != nil {
		return err
	}
	zhLangs := sole(zh.Main).LocaleDisplayNames.Languages
	enLangs := sole(en.Main).LocaleDisplayNames.Languages

	isTwoLetter := func(s string) bool {
		if len(s) != 2 {
			return false
		}
		return s[0] >= 'a' && s[0] <= 'z' && s[1] >= 'a' && s[1] <= 'z'
	}

	tags := map[string]bool{"zh-Hans": true, "zh-Hant": true}
	for tag := range zhLangs {
		if isTwoLetter(tag) && enLangs[tag] != "" && zhLangs[tag] != "" {
			tags[tag] = true
		}
	}

	names := make(map[string]map[string]string, len(locales))
	for _, l := range locales {
		var lf languagesFile
		if l == "zh-Hans" {
			names[l] = map[string]string{}
			continue // filled from the already-fetched zh table below
		}
		if err := getJSON(urlLanguages(l), cacheDir, &lf); err != nil {
			return fmt.Errorf("languages %s: %w", l, err)
		}
		names[l] = sole(lf.Main).LocaleDisplayNames.Languages
	}
	names["zh-Hans"] = zhLangs

	// Keep only the selected tags; require a name in every locale, falling
	// back to the English name when a locale lacks the entry.
	filtered := make(map[string]map[string]string, len(locales))
	for _, l := range locales {
		filtered[l] = make(map[string]string, len(tags))
		for tag := range tags {
			n := names[l][tag]
			if n == "" {
				n = enLangs[tag]
			}
			if n == "" {
				n = tag
			}
			filtered[l][tag] = n
		}
	}

	orders := make(map[string][]string, len(locales))
	for _, l := range locales {
		orders[l] = collatedOrder(l, filtered[l])
	}

	// Country -> official languages (official + de facto official, most
	// spoken first), restricted to the directory's tag set.
	var tif territoryInfoFile
	if err := getJSON(urlTerritoryInfo, cacheDir, &tif); err != nil {
		return err
	}
	type langPop struct {
		tag string
		pct float64
	}
	countryLangs := make(map[string][]string)
	for cc, info := range tif.Supplemental.TerritoryInfo {
		var picks, byPopulation []langPop
		for tag, lp := range info.LanguagePopulation {
			if !tags[tag] {
				continue // not in the served language directory
			}
			pct, _ := strconv.ParseFloat(lp.PopulationPercent, 64)
			if lp.OfficialStatus == "official" || lp.OfficialStatus == "de_facto_official" {
				picks = append(picks, langPop{tag, pct})
			}
			// official_regional / minority entries only ever surface via
			// the no-official fallback below, never on their own.
			byPopulation = append(byPopulation, langPop{tag, pct})
		}
		// Fallback for territories whose languagePopulation entries carry
		// no _officialStatus at all: CLDR territoryInfo lists Ascension
		// Island and Tristan da Cunha as 99% English but leaves the status
		// attribute unset, so the official-only filter above would hand
		// both an empty language list and GetCountryDefaults no answer.
		// When nothing official was found, fall back to the plain
		// population ranking (AC/TA -> en). Territories with no
		// languagePopulation data at all still yield an empty list.
		if len(picks) == 0 {
			picks = byPopulation
		}
		sort.Slice(picks, func(i, j int) bool {
			if picks[i].pct != picks[j].pct {
				return picks[i].pct > picks[j].pct
			}
			return picks[i].tag < picks[j].tag
		})
		if len(picks) > 0 {
			out := make([]string, len(picks))
			for i, p := range picks {
				out[i] = p.tag
			}
			countryLangs[cc] = out
		}
	}

	// Native names (endonyms): each language named in its own script —
	// 日本語, 한국어, English — sourced from the language's OWN CLDR file.
	// Best-effort: languages without a CLDR locale stay empty and the UI
	// simply omits the column.
	natives := make(map[string]string, len(tags))
	for _, tag := range sortedKeys(tags) {
		var lf languagesFile
		if err := getJSON(urlLanguages(tag), cacheDir, &lf); err != nil {
			continue // no CLDR locale for this tag — no endonym
		}
		if n := sole(lf.Main).LocaleDisplayNames.Languages[tag]; n != "" {
			natives[tag] = n
		}
	}

	w := newWriter("CLDR " + cldrVersion + " languages (639-1 set + zh-Hans/zh-Hant) + territoryInfo official languages + endonyms")
	w.line("// Languages is the selectable BCP 47 tag table keyed by tag;")
	w.line("// NativeName is the endonym (the language's own name for itself).")
	w.line("var Languages = map[string]Language{")
	for _, tag := range sortedKeys(tags) {
		w.line("\t%q: {Tag: %q, NativeName: %q},", tag, tag, natives[tag])
	}
	w.line("}")
	w.line("")
	w.emitNames("LanguageNames", filtered)
	w.emitOrder("LanguageOrder", orders)
	w.line("// CountryLanguages maps alpha-2 -> official languages (official and")
	w.line("// de facto official only, most-spoken first; territories with no")
	w.line("// official-status entry fall back to plain population ranking —")
	w.line("// e.g. AC/TA get en from CLDR's 99% share without a status flag).")
	w.line("var CountryLanguages = map[string][]string{")
	for _, cc := range sortedKeys(countryLangs) {
		w.line("\t%q: {%s},", cc, quoteJoin(countryLangs[cc]))
	}
	w.line("}")
	w.line("")
	return w.save(outputPath("languages_data.go"))
}
