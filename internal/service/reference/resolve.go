// ResolveCodes: the batch by-code lookup across domains. One
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

// ResolveCodes batch-resolves codes across domains; misses land in missing_*.
func (*Service) ResolveCodes(_ context.Context, req *pb.ResolveCodesRequest) (*pb.ResolveCodesResponse, error) {
	l := resolveLocale(req.GetLocale())
	resp := &pb.ResolveCodesResponse{DataVersion: data.Version}

	for _, code := range req.GetRegionCodes() {
		if c, ok := data.Countries[code]; ok {
			resp.Countries = append(resp.Countries, countryRow(l, code, c))
		} else {
			resp.MissingRegions = append(resp.MissingRegions, code)
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
				Id:          t.ID,
				Aliases:     t.Aliases,
				RegionCodes: t.RegionCodes,
				Name:        data.TimezoneNames[l][canonical],
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
				Tag:        key,
				Name:       data.LanguageNames[l][key],
				NativeName: data.Languages[key].NativeName,
			})
		} else {
			resp.MissingLanguages = append(resp.MissingLanguages, tag)
		}
	}

	for _, code := range req.GetCurrencyCodes() {
		if c, ok := data.Currencies[code]; ok {
			resp.Currencies = append(resp.Currencies, &pb.Currency{
				Code:        c.Code,
				Symbol:      c.Symbol,
				MinorUnits:  c.MinorUnits,
				RegionCodes: c.RegionCodes,
				Name:        data.CurrencyNames[l][code],
				FlagEmoji:   c.FlagEmoji,
			})
		} else {
			resp.MissingCurrencies = append(resp.MissingCurrencies, code)
		}
	}
	return resp, nil
}
