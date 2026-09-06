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
}

func buildCurrencies(cacheDir string) error {
	var cd currencyDataFile
	if err := getJSON(urlCurrencyData, cacheDir, &cd); err != nil {
		return err
	}
	fractions := cd.Supplemental.CurrencyData.Fractions
	regions := cd.Supplemental.CurrencyData.Region

	names := make(map[string]map[string]currencyNameEntry, len(locales))
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
		entities[code] = currencyRow{Code: code, MinorUnits: digits, CountryCodes: regs}
	}

	// Names + symbols; every locale names every code, falling back to
	// English, then to the code itself for both.
	nameTables := make(map[string]map[string]string, len(locales))
	for _, l := range locales {
		nameTables[l] = make(map[string]string, len(entities))
		for code := range entities {
			n := names[l][code].DisplayName
			if n == "" {
				n = names["en"][code].DisplayName
			}
			if n == "" {
				n = code
			}
			nameTables[l][code] = n
		}
	}
	for code, row := range entities {
		if row.Symbol == "" {
			row.Symbol = firstNonEmpty(names["en"][code].Symbol, code)
			entities[code] = row
		}
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
		w.line("\t%q: {Code: %q, Symbol: %q, MinorUnits: %d, CountryCodes: []string{%s}},",
			code, e.Code, e.Symbol, e.MinorUnits, quoteJoin(e.CountryCodes))
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
