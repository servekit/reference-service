// Package reference serves the compiled reference tables. Every method is
// a pure function over internal/data — no state, no IO — so the module-mode
// Handler costs nothing to embed. Locale handling lives here once for all
// domains: requests carry an optional BCP 47 tag, empty defaults to
// zh-Hans, unresolvable tags fall back to en, and the locale never fails
// the call.
package reference

import (
	"context"

	"golang.org/x/text/language"

	pb "github.com/servekit/api/gen/go/reference/v1"

	"github.com/servekit/reference-service/internal/data"
)

var tags = func() []language.Tag {
	ts := make([]language.Tag, len(data.Locales))
	for i, l := range data.Locales {
		ts[i] = language.MustParse(l)
	}
	return ts
}()

var matcher = language.NewMatcher(tags)

// resolveLocale maps a request locale to a compiled key: "" -> zh-Hans;
// syntactically broken, empty-language, or unmatchable tags -> "en";
// otherwise the matcher's pick (zh-TW -> zh-Hant, pt-BR -> pt) via
// likelySubtags expansion.
func resolveLocale(locale string) string {
	if locale == "" {
		return data.Locales[0]
	}
	tag, err := language.Parse(locale)
	if err != nil || tag == language.Und {
		return "en"
	}
	_, idx, conf := matcher.Match(tag)
	if conf == language.No {
		return "en"
	}
	return data.Locales[idx]
}

// Service implements the reference domain over the compiled tables.
type Service struct{}

// New constructs the reference domain service.
func New() *Service { return &Service{} }

// ListCountries returns the full country directory in the request locale.
func (*Service) ListCountries(_ context.Context, req *pb.ListCountriesRequest) (*pb.ListCountriesResponse, error) {
	l := resolveLocale(req.GetLocale())
	out := make([]*pb.Country, 0, len(data.CountryOrder[l]))
	for _, code := range data.CountryOrder[l] {
		out = append(out, countryRow(l, code, data.Countries[code]))
	}
	return &pb.ListCountriesResponse{Countries: out, DataVersion: data.Version}, nil
}

// countryRow builds the wire Country for one alpha-2 in locale l — the
// single source every country-returning handler shares, so ListCountries,
// GetCountries, and the per-code lookups can never drift apart.
func countryRow(l, code string, c data.Country) *pb.Country {
	return &pb.Country{
		RegionCode:    c.Code,
		Alpha_3:       c.Alpha3,
		DialCode:      c.DialCode,
		FlagEmoji:     c.FlagEmoji,
		Name:          data.CountryNames[l][code],
		ExampleNumber: c.ExampleNumber,
		LanguageTags:  data.CountryLanguages[code],
	}
}

// GetCountries returns the directory rows for exactly the requested
// alpha-2 codes, in request order — the subset-fetch counterpart to
// ListCountries' full collated listing. Unknown codes land in
// missing_regions (request order) and never fail the call, so a caller
// can validate its served-country config in the same round-trip.
func (*Service) GetCountries(_ context.Context, req *pb.GetCountriesRequest) (*pb.GetCountriesResponse, error) {
	l := resolveLocale(req.GetLocale())
	resp := &pb.GetCountriesResponse{DataVersion: data.Version}
	for _, code := range req.GetRegionCodes() {
		if c, ok := data.Countries[code]; ok {
			resp.Countries = append(resp.Countries, countryRow(l, code, c))
		} else {
			resp.MissingRegions = append(resp.MissingRegions, code)
		}
	}
	return resp, nil
}

// ListTimezones returns the canonical IANA zones in the request locale.
func (*Service) ListTimezones(_ context.Context, req *pb.ListTimezonesRequest) (*pb.ListTimezonesResponse, error) {
	l := resolveLocale(req.GetLocale())
	names := data.TimezoneNames[l]
	out := make([]*pb.Timezone, 0, len(data.TimezoneOrder[l]))
	for _, id := range data.TimezoneOrder[l] {
		t := data.Timezones[id]
		out = append(out, &pb.Timezone{
			Id:          t.ID,
			Aliases:     t.Aliases,
			RegionCodes: t.RegionCodes,
			Name:        names[id],
		})
	}
	return &pb.ListTimezonesResponse{Timezones: out, DataVersion: data.Version}, nil
}

// ListLanguages returns the selectable BCP 47 tags in the request locale.
func (*Service) ListLanguages(_ context.Context, req *pb.ListLanguagesRequest) (*pb.ListLanguagesResponse, error) {
	l := resolveLocale(req.GetLocale())
	names := data.LanguageNames[l]
	out := make([]*pb.Language, 0, len(data.LanguageOrder[l]))
	for _, tag := range data.LanguageOrder[l] {
		out = append(out, &pb.Language{
			Tag:        tag,
			Name:       names[tag],
			NativeName: data.Languages[tag].NativeName,
		})
	}
	return &pb.ListLanguagesResponse{Languages: out, DataVersion: data.Version}, nil
}

// ListCurrencies returns the ISO 4217 active set in the request locale.
func (*Service) ListCurrencies(_ context.Context, req *pb.ListCurrenciesRequest) (*pb.ListCurrenciesResponse, error) {
	l := resolveLocale(req.GetLocale())
	names := data.CurrencyNames[l]
	out := make([]*pb.Currency, 0, len(data.CurrencyOrder[l]))
	for _, code := range data.CurrencyOrder[l] {
		c := data.Currencies[code]
		out = append(out, &pb.Currency{
			Code:        c.Code,
			Symbol:      c.Symbol,
			MinorUnits:  c.MinorUnits,
			RegionCodes: c.RegionCodes,
			Name:        names[code],
			FlagEmoji:   c.FlagEmoji,
		})
	}
	return &pb.ListCurrenciesResponse{Currencies: out, DataVersion: data.Version}, nil
}

// ListRegionGroups returns the UN M49 hierarchy (continents first).
func (*Service) ListRegionGroups(_ context.Context, req *pb.ListRegionGroupsRequest) (*pb.ListRegionGroupsResponse, error) {
	l := resolveLocale(req.GetLocale())
	names := data.RegionGroupNames[l]
	out := make([]*pb.RegionGroup, 0, len(data.RegionGroupOrder[l]))
	for _, code := range data.RegionGroupOrder[l] {
		g := data.RegionGroups[code]
		out = append(out, &pb.RegionGroup{
			GroupCode:   g.Code,
			ParentCode:  g.ParentCode,
			RegionCodes: g.RegionCodes,
			Name:        names[code],
		})
	}
	return &pb.ListRegionGroupsResponse{RegionGroups: out, DataVersion: data.Version}, nil
}
