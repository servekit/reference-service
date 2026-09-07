// cldr.go — upstream URL matrix and the JSON shapes actually consumed.
// CLDR ships as split npm packages on unpkg; everything is pinned to 45.0.0.
package main

import (
	"encoding/json"
	"fmt"
	"strings"
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
	urlTerritoryInfo        = pkgCore + "/supplemental/territoryInfo.json"
	urlPrimaryZones         = pkgCore + "/supplemental/primaryZones.json"
	urlMetaZones            = pkgCore + "/supplemental/metaZones.json"
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

// timezoneNamesFile: CLDR nests zone trees arbitrarily deep
// (zone > America > Argentina > Buenos_Aires), so the tree is decoded as
// generic maps and walked segment by segment. The metazone section carries
// the locale's names for zone groups (Argentina Time, ...).
type timezoneNamesFile struct {
	Main map[string]struct {
		Dates struct {
			TimeZoneNames struct {
				Zone     map[string]any `json:"zone"`
				Metazone map[string]struct {
					Long struct {
						Generic  string `json:"generic"`
						Standard string `json:"standard"`
					} `json:"long"`
				} `json:"metazone"`
			} `json:"timeZoneNames"`
		} `json:"dates"`
	} `json:"main"`
}

// metaZonesFile models CLDR supplemental metaZones: the territory-001
// mapZone list is a flat zone-id -> metazone table, and metazoneInfo is the
// per-zone history tree (zone -> dated metazone intervals; the entry with
// no _to is current). Both spell zone ids with their historical names —
// callers normalize via the backward-link table.
type metaZonesFile struct {
	Supplemental struct {
		MetaZones struct {
			Metazones []struct {
				MapZone struct {
					Type      string `json:"_type"`
					Territory string `json:"_territory"`
					Other     string `json:"_other"`
				} `json:"mapZone"`
			} `json:"metazones"`
			MetazoneInfo map[string]any `json:"metazoneInfo"`
		} `json:"metaZones"`
	} `json:"supplemental"`
}

// currentMetazone extracts the usesMetazone entry with no _to (the zone's
// current metazone) from a metazoneInfo leaf list.
func currentMetazone(leaf any) string {
	list, ok := leaf.([]any)
	if !ok {
		return ""
	}
	for _, e := range list {
		uses, ok := e.(map[string]any)["usesMetazone"].(map[string]any)
		if !ok {
			continue
		}
		if _, hasTo := uses["_to"]; !hasTo {
			if m, ok := uses["_mzone"].(string); ok {
				return m
			}
		}
	}
	return ""
}

// walkMetazoneInfo visits every zone leaf under the timezone tree and calls
// fn(zoneID, metazone) for entries with a current metazone. A leaf is a
// list of dated usesMetazone intervals; interior nodes are dicts.
func walkMetazoneInfo(node map[string]any, prefix string, fn func(zone, mzone string)) {
	for k, v := range node {
		zone := k
		if prefix != "" {
			zone = prefix + "/" + k
		}
		if _, isList := v.([]any); isList {
			if m := currentMetazone(v); m != "" {
				fn(zone, m)
			}
			continue
		}
		if children, isDict := v.(map[string]any); isDict {
			walkMetazoneInfo(children, zone, fn)
		}
	}
}

// exemplarCity walks a zone tree along the id's segments and returns the
// deepest exemplarCity ("" when the path is absent).
func exemplarCity(tree map[string]any, id string) string {
	segs := strings.Split(id, "/")
	node := tree
	for i, seg := range segs {
		next, ok := node[seg].(map[string]any)
		if !ok {
			_ = i
			return ""
		}
		node = next
	}
	if city, ok := node["exemplarCity"].(string); ok {
		return city
	}
	return ""
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

// territoryInfoFile models CLDR supplemental territoryInfo: per-territory
// language population with official-status annotations.
type territoryInfoFile struct {
	Supplemental struct {
		TerritoryInfo map[string]struct {
			LanguagePopulation map[string]struct {
				PopulationPercent string `json:"_populationPercent"`
				OfficialStatus    string `json:"_officialStatus"`
			} `json:"languagePopulation"`
		} `json:"territoryInfo"`
	} `json:"supplemental"`
}

// primaryZonesFile models CLDR supplemental primaryZones — the handful of
// countries where CLDR's primary zone differs from zone1970.tab's
// first-row-for-country rule.
type primaryZonesFile struct {
	Supplemental struct {
		PrimaryZones map[string]string `json:"primaryZones"`
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
