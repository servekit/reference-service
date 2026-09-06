// cldr.go — upstream URL matrix and the JSON shapes actually consumed.
// CLDR ships as split npm packages on unpkg; everything is pinned to 45.0.0.
package main

import (
	"encoding/json"
	"fmt"
)

const cldrVersion = "45.0.0"

const (
	pkgLocalenames = "https://unpkg.com/cldr-localenames-modern@" + cldrVersion
	pkgDates       = "https://unpkg.com/cldr-dates-modern@" + cldrVersion
	pkgNumbers     = "https://unpkg.com/cldr-numbers-modern@" + cldrVersion
	pkgCore        = "https://unpkg.com/cldr-core@" + cldrVersion

	urlISO3166 = "https://raw.githubusercontent.com/lukes/ISO-3166-Countries-with-Regional-Codes/master/all/all.json"
	urlZoneTab = "https://raw.githubusercontent.com/eggert/tz/main/zone1970.tab"
	urlTzBack  = "https://raw.githubusercontent.com/eggert/tz/main/backward"
)

func urlTerritories(locale string) string {
	return pkgLocalenames + "/main/" + locale + "/territories.json"
}

func urlLanguages(locale string) string {
	return pkgLocalenames + "/main/" + locale + "/languages.json"
}

func urlTimezoneNames(locale string) string {
	return pkgDates + "/main/" + locale + "/timeZoneNames.json"
}

func urlCurrencies(locale string) string {
	return pkgNumbers + "/main/" + locale + "/currencies.json"
}

const (
	urlTerritoryContainment = pkgCore + "/supplemental/territoryContainment.json"
	urlCurrencyData         = pkgCore + "/supplemental/currencyData.json"
)

// CLDR wraps every main/<locale> file in main.<locale> plus a level keyed by
// content; only the leaves we need are modeled here.

type territoriesFile struct {
	Main map[string]struct {
		LocaleDisplayNames struct {
			Territories map[string]string `json:"territories"`
		} `json:"localeDisplayNames"`
	} `json:"main"`
}

// sole returns the single value of a main/<locale> wrapper map.
func sole[T any](m map[string]T) T {
	var zero T
	for _, v := range m {
		return v
	}
	return zero
}

type languagesFile struct {
	Main map[string]struct {
		LocaleDisplayNames struct {
			Languages map[string]string `json:"languages"`
		} `json:"localeDisplayNames"`
	} `json:"main"`
}

type timezoneNamesFile struct {
	Main map[string]struct {
		Dates struct {
			TimeZoneNames struct {
				Zone map[string]map[string]struct {
					ExemplarCity string `json:"exemplarCity"`
				} `json:"zone"`
			} `json:"timeZoneNames"`
		} `json:"dates"`
	} `json:"main"`
}

// currencyNameEntry is one CLDR currencies.json leaf.
type currencyNameEntry struct {
	DisplayName string `json:"displayName"`
	Symbol      string `json:"symbol"`
}

type currenciesFile struct {
	Main map[string]struct {
		Numbers struct {
			Currencies map[string]currencyNameEntry `json:"currencies"`
		} `json:"numbers"`
	} `json:"main"`
}

type territoryContainmentFile struct {
	Supplemental struct {
		TerritoryContainment map[string]struct {
			Contains []string `json:"_contains"`
			Status   string   `json:"_status"`
		} `json:"territoryContainment"`
	} `json:"supplemental"`
}

type currencyDataFile struct {
	Supplemental struct {
		CurrencyData struct {
			Fractions map[string]map[string]any `json:"fractions"`
			Region    map[string][]map[string]struct {
				From string `json:"_from"`
				To   string `json:"_to"`
			} `json:"region"`
		} `json:"currencyData"`
	} `json:"supplemental"`
}

type isoRow struct {
	Name    string `json:"name"`
	Alpha2  string `json:"alpha-2"`
	Alpha3  string `json:"alpha-3"`
	Region  string `json:"region"`
	SubRegn string `json:"sub-region"`
}

// getJSON fetches url and unmarshals into out.
func getJSON(url, cacheDir string, out any) error {
	b, err := fetch(url, cacheDir)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(b, out); err != nil {
		return fmt.Errorf("parse %s: %w", url, err)
	}
	return nil
}
