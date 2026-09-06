// resolve.go — ResolveCodes: the batch by-code lookup across domains. One
// call serves a registration form's validate-and-fetch; unknown codes land
// in missing_* and never fail the call. Timezone ids accept backward
// aliases (normalized to canonical); language tags are case-canonicalized
// via x/text before lookup.
package reference

import (
	"context"

	"golang.org/x/text/language"

	pb "github.com/servekit/api/gen/go/reference/v1"
	"github.com/servekit/reference-service/internal/data"
)

func (s *Service) ResolveCodes(_ context.Context, req *pb.ResolveCodesRequest) (*pb.ResolveCodesResponse, error) {
	l := resolveLocale(req.GetLocale())
	resp := &pb.ResolveCodesResponse{DataVersion: data.Version}

	for _, code := range req.GetCountryCodes() {
		if c, ok := data.Countries[code]; ok {
			resp.Countries = append(resp.Countries, &pb.Country{
				Code:      c.Code,
				Alpha_3:   c.Alpha3,
				DialCode:  c.DialCode,
				FlagEmoji: c.FlagEmoji,
				Name:      data.CountryNames[l][code],
			})
		} else {
			resp.MissingCountries = append(resp.MissingCountries, code)
		}
	}

	for _, id := range req.GetTimezoneIds() {
		if id == "" {
			resp.MissingTimezones = append(resp.MissingTimezones, id)
			continue
		}
		canonical := id
		if _, ok := data.Timezones[canonical]; !ok {
			if target, ok := data.TimezoneAliases[canonical]; ok {
				canonical = target
			}
		}
		if t, ok := data.Timezones[canonical]; ok {
			resp.Timezones = append(resp.Timezones, &pb.Timezone{
				Id:           t.ID,
				Aliases:      t.Aliases,
				CountryCodes: t.CountryCodes,
				Name:         data.TimezoneNames[l][canonical],
			})
		} else {
			resp.MissingTimezones = append(resp.MissingTimezones, id)
		}
	}

	for _, tag := range req.GetLanguageTags() {
		t, err := language.Parse(tag)
		if err != nil {
			resp.MissingLanguages = append(resp.MissingLanguages, tag)
			continue
		}
		key := t.String() // canonical case: "ZH-HANS" -> "zh-Hans"
		if _, ok := data.Languages[key]; ok {
			resp.Languages = append(resp.Languages, &pb.Language{
				Tag:  key,
				Name: data.LanguageNames[l][key],
			})
		} else {
			resp.MissingLanguages = append(resp.MissingLanguages, tag)
		}
	}

	for _, code := range req.GetCurrencyCodes() {
		if c, ok := data.Currencies[code]; ok {
			resp.Currencies = append(resp.Currencies, &pb.Currency{
				Code:         c.Code,
				Symbol:       c.Symbol,
				MinorUnits:   c.MinorUnits,
				CountryCodes: c.CountryCodes,
				Name:         data.CurrencyNames[l][code],
			})
		} else {
			resp.MissingCurrencies = append(resp.MissingCurrencies, code)
		}
	}
	return resp, nil
}
