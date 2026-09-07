// Package data holds the compiled reference tables. The generator that
// produces the *_data.go files lives in tools/gen; everything here is
// immutable at runtime — no DB, no config, data ships with the binary.
//
// Layout per domain: an entity map keyed by code/id/tag, a Names table
// (locale -> code -> display name), and an Order table (locale -> codes in
// that locale's collation order) — the List handlers walk Order and join
// with Names + entity map. Locales[0] is the request-default locale.
package data

// Version is the data snapshot stamp (generation date), shared by every
// domain and returned as data_version on each response.
const Version = "2026-09-07"

// Locales lists the compiled locales in priority order; Locales[0] (zh-Hans)
// is the default when a request omits locale. Unresolvable locales fall
// back to "en".
var Locales = []string{
	"zh-Hans", "zh-Hant", "ja", "ko", "en", "fr", "de", "pt", "es", "ar", "ru",
}

// Country is one row of the country directory (locale-independent fields).
//
// Alpha3 provenance for the three non-officially-assigned regions this
// directory serves: AC=ASC and TA=TAA are alpha-3 code elements the ISO
// 3166 reserved list holds at the UPU's request (also used by ITU; the
// UPU's GB addressing documentation lists both), and XK=XKX is the
// user-assigned alpha-3 for Kosovo adopted de facto by the European
// Commission, SWIFT, and the World Bank. Sources and overlay logic live
// in tools/gen/countries.go (reservedAlpha3).
type Country struct {
	Code          string // ISO 3166-1 alpha-2
	Alpha3        string // ISO 3166-1 alpha-3 (reserved/de-facto value for AC/TA/XK — see type doc)
	DialCode      string // ITU E.164 with "+", e.g. "+86"
	FlagEmoji     string // regional-indicator pair, e.g. "🇨🇳"
	ExampleNumber string // libphonenumber example, e.g. "+86 138 0013 8000"; "" if none
}

// Timezone is one canonical IANA zone.
type Timezone struct {
	ID          string   // canonical, e.g. "Asia/Shanghai"
	Aliases     []string // tzdb backward links resolved to this zone
	RegionCodes []string // alpha-2 members from zone1970.tab
}

// Language is one selectable BCP 47 tag.
type Language struct {
	Tag        string
	NativeName string // endonym, e.g. "日本語"; "" when CLDR lacks the locale
}

// Currency is one ISO 4217 currency.
type Currency struct {
	Code        string
	Symbol      string
	MinorUnits  int32
	RegionCodes []string // regions where currently valid
	FlagEmoji   string   // issuer flag; "" when no country backs the code (XDR)
}

// RegionGroup is one UN M49 node; top level (ParentCode == "") is a
// continent — the world root (001) is not served.
type RegionGroup struct {
	Code        string
	ParentCode  string
	RegionCodes []string // direct members only
}
