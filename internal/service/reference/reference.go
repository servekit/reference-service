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
	names := data.CountryNames[l]
	out := make([]*pb.Country, 0, len(data.CountryOrder[l]))
	for _, code := range data.CountryOrder[l] {
		c := data.Countries[code]
		out = append(out, &pb.Country{
			Code:          c.Code,
			Alpha_3:       c.Alpha3,
			DialCode:      c.DialCode,
			FlagEmoji:     c.FlagEmoji,
			Name:          names[code],
			ExampleNumber: c.ExampleNumber,
		})
	}
	return &pb.ListCountriesResponse{Countries: out, DataVersion: data.Version}, nil
}

// ListTimezones returns the canonical IANA zones in the request locale.
func (*Service) ListTimezones(_ context.Context, req *pb.ListTimezonesRequest) (*pb.ListTimezonesResponse, error) {
	l := resolveLocale(req.GetLocale())
	names := data.TimezoneNames[l]
	out := make([]*pb.Timezone, 0, len(data.TimezoneOrder[l]))
	for _, id := range data.TimezoneOrder[l] {
		t := data.Timezones[id]
		out = append(out, &pb.Timezone{
			Id:           t.ID,
			Aliases:      t.Aliases,
			CountryCodes: t.CountryCodes,
			Name:         names[id],
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
		out = append(out, &pb.Language{Tag: tag, Name: names[tag]})
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
			Code:         c.Code,
			Symbol:       c.Symbol,
			MinorUnits:   c.MinorUnits,
			CountryCodes: c.CountryCodes,
			Name:         names[code],
			FlagEmoji:    c.FlagEmoji,
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
			Code:         g.Code,
			ParentCode:   g.ParentCode,
			CountryCodes: g.CountryCodes,
			Name:         names[code],
		})
	}
	return &pb.ListRegionGroupsResponse{RegionGroups: out, DataVersion: data.Version}, nil
}
