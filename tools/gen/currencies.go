// currencies.go — builds the currency domain: the ISO 4217 active set and
// country membership from CLDR region validity ranges (currently-valid
// entries only), minor units from currency fractions (DEFAULT when absent),
// and names/symbols from each locale's CLDR currencies table.
package main

import (
	"fmt"
	"sort"
	"strconv"
)

type currencyRow struct {
	Code         string
	Symbol       string
	MinorUnits   int32
	CountryCodes []string
	FlagEmoji    string
}

// Issuer-territory knowledge for flag derivation, assembled once in
// buildCurrencies: the English CLDR territory table (covers
// exceptionally-reserved issuers like EU) plus the dialing-region set.
var (
	enTerritoryNames map[string]string
	dialRegions      map[string]string
)

func buildCurrencies(cacheDir string) error {
	var cd currencyDataFile
	if err := getJSON(urlCurrencyData, cacheDir, &cd); err != nil {
		return err
	}
	var tf territoriesFile
	if err := getJSON(urlTerritories("en"), cacheDir, &tf); err != nil {
		return err
	}
	enTerritoryNames = sole(tf.Main).LocaleDisplayNames.Territories
	dialRegions = derivedDials()
	fractions := cd.Supplemental.CurrencyData.Fractions
	regions := cd.Supplemental.CurrencyData.Region

	names := make(map[string]map[string]map[string]string, len(locales))
	for _, l := range locales {
		var cf currenciesFile
		if err := getJSON(urlCurrencies(l), cacheDir, &cf); err != nil {
			return fmt.Errorf("currencies %s: %w", l, err)
		}
		names[l] = sole(cf.Main).Numbers.Currencies
	}

	// Active set: currencies with a currently-valid (_from, no _to) span in
	// some region's history. fractions only lists NON-default minor units
	// (plus DEFAULT), so it cannot define the code set.
	type regionUse struct {
		code string
		regs []string
	}
	active := make(map[string]*regionUse)
	for region, entries := range regions {
		for _, e := range entries {
			for code, span := range e {
				if span.From == "" || span.To != "" {
					continue
				}
				u, ok := active[code]
				if !ok {
					u = &regionUse{code: code}
					active[code] = u
				}
				u.regs = append(u.regs, region)
			}
		}
	}

	defaultDigits := int32(2)
	if d, ok := fractions["DEFAULT"]["_digits"].(string); ok {
		if n, err := strconv.Atoi(d); err == nil {
			defaultDigits = int32(n)
		}
	}

	entities := make(map[string]currencyRow, len(active))
	for code, u := range active {
		digits := defaultDigits
		if frac, ok := fractions[code]; ok {
			if d, ok := frac["_digits"].(string); ok {
				if n, err := strconv.Atoi(d); err == nil {
					digits = int32(n)
				}
			}
		}
		regs := u.regs
		sort.Strings(regs)
		entities[code] = currencyRow{
			Code:         code,
			MinorUnits:   digits,
			CountryCodes: regs,
			FlagEmoji:    currencyFlag(code, regs),
		}
	}

	// Names + symbols; every locale names every code, falling back to
	// English, then to the code itself.
	nameTables := make(map[string]map[string]string, len(locales))
	for _, l := range locales {
		nameTables[l] = make(map[string]string, len(entities))
		for code := range entities {
			n := names[l][code]["displayName"]
			if n == "" {
				n = names["en"][code]["displayName"]
			}
			if n == "" {
				n = code
			}
			nameTables[l][code] = n
		}
	}
	// Symbol: the NARROW variant first ("¥", "$", "€") — one glyph, the
	// picker/form display everyone expects; CLDR's standard symbol is the
	// disambiguated form (CN¥/A$) which only pays off in mixed-currency
	// tables, and our rows already carry the code for that.
	for code, row := range entities {
		row.Symbol = firstNonEmpty(
			names["en"][code]["symbol-alt-narrow"],
			names["en"][code]["symbol"],
			code,
		)
		entities[code] = row
	}

	orders := make(map[string][]string, len(locales))
	for _, l := range locales {
		orders[l] = collatedOrder(l, nameTables[l])
	}

	w := newWriter("CLDR " + cldrVersion + " currencyData (fractions + region validity) + currencies names")
	w.line("// Currencies is the ISO 4217 active table keyed by code.")
	w.line("var Currencies = map[string]Currency{")
	for _, code := range sortedKeys(entities) {
		e := entities[code]
		w.line("\t%q: {Code: %q, Symbol: %q, MinorUnits: %d, CountryCodes: []string{%s}, FlagEmoji: %q},",
			code, e.Code, e.Symbol, e.MinorUnits, quoteJoin(e.CountryCodes), e.FlagEmoji)
	}
	w.line("}")
	w.line("")
	w.emitNames("CurrencyNames", nameTables)
	w.emitOrder("CurrencyOrder", orders)
	return w.save(outputPath("currencies_data.go"))
}

func firstNonEmpty(ss ...string) string {
	for _, s := range ss {
		if s != "" {
			return s
		}
	}
	return ""
}

// currencyFlag derives the issuer flag. ISO 4217 alpha-3 codes embed the
// issuing authority's alpha-2 in their first two letters (CNY -> CN, EUR ->
// the exceptionally-reserved EU, which renders as the EU flag); that pair
// is the flag when it is a known territory. X-prefixed multi-country codes
// (XOF/XAF/XCD/XPF) take their first member; country-less codes (XDR, the
// metal units) stay empty.
func currencyFlag(code string, members []string) string {
	head := code[:2]
	if head == "EU" || dialRegions[head] != "" ||
		enTerritoryNames[head] != "" || enTerritoryNames[head+"-alt-short"] != "" {
		return flagEmoji(head)
	}
	if len(members) > 0 {
		return flagEmoji(members[0])
	}
	return ""
}
