// languages.go — builds the language domain: the selectable BCP 47 tag set
// is the two-letter (ISO 639-1) subset that zh-Hans and en both name, plus
// the script-qualified tags we serve (zh-Hans, zh-Hant); names come from
// each locale's CLDR languages table.
package main

import "fmt"

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

	w := newWriter("CLDR " + cldrVersion + " languages (639-1 set + zh-Hans/zh-Hant)")
	w.line("// Languages is the selectable BCP 47 tag table keyed by tag.")
	w.line("var Languages = map[string]Language{")
	for _, tag := range sortedKeys(tags) {
		w.line("\t%q: {Tag: %q},", tag, tag)
	}
	w.line("}")
	w.line("")
	w.emitNames("LanguageNames", filtered)
	w.emitOrder("LanguageOrder", orders)
	return w.save(outputPath("languages_data.go"))
}
