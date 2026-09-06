// profile.go — GetCountryProfile: aggregates the compiled domains for one
// country. A pure join, no new sources: base info from Countries, the
// continent/sub-region chain from RegionGroups (direct membership plus
// ancestors, continent first), official languages from CountryLanguages,
// and the country's currently-valid currencies and IANA zones. Lists keep
// each domain's per-locale collation order.

package reference

import (
	"context"

	pb "github.com/servekit/api/gen/go/reference/v1"

	"github.com/servekit/reference-service/internal/data"
	"github.com/servekit/reference-service/pkg/xcodes"
)

// GetCountryProfile returns everything the domains know about one country.
func (*Service) GetCountryProfile(_ context.Context, req *pb.GetCountryProfileRequest) (*pb.GetCountryProfileResponse, error) {
	code := req.GetCountryCode()
	c, ok := data.Countries[code]
	if !ok {
		return nil, xcodes.ErrCountryNotFound.Wrapf(nil, "country %s", code)
	}
	l := resolveLocale(req.GetLocale())

	// Region chain: groups the country is a direct member of, plus their
	// ancestors. Collect the set first, then emit in RegionGroupOrder —
	// that order is top-down, so the continent lands before sub-regions.
	chainSet := map[string]bool{}
	for gCode, g := range data.RegionGroups {
		if containsStr(g.CountryCodes, code) {
			for cur := gCode; cur != ""; cur = data.RegionGroups[cur].ParentCode {
				chainSet[cur] = true
			}
		}
	}
	var chain []string
	for _, g := range data.RegionGroupOrder[l] {
		if chainSet[g] {
			chain = append(chain, g)
		}
	}

	resp := &pb.GetCountryProfileResponse{
		Country: &pb.Country{
			Code:      c.Code,
			Alpha_3:   c.Alpha3,
			DialCode:  c.DialCode,
			FlagEmoji: c.FlagEmoji,
			Name:      data.CountryNames[l][code],
		},
		DataVersion: data.Version,
	}
	for _, g := range chain {
		rg := data.RegionGroups[g]
		resp.RegionGroups = append(resp.RegionGroups, &pb.RegionGroup{
			Code:         rg.Code,
			ParentCode:   rg.ParentCode,
			CountryCodes: rg.CountryCodes,
			Name:         data.RegionGroupNames[l][g],
		})
	}
	for _, tag := range data.CountryLanguages[code] {
		resp.Languages = append(resp.Languages, &pb.Language{
			Tag:  tag,
			Name: data.LanguageNames[l][tag],
		})
	}
	for _, cc := range data.CurrencyOrder[l] {
		cur := data.Currencies[cc]
		if containsStr(cur.CountryCodes, code) {
			resp.Currencies = append(resp.Currencies, &pb.Currency{
				Code:         cur.Code,
				Symbol:       cur.Symbol,
				MinorUnits:   cur.MinorUnits,
				CountryCodes: cur.CountryCodes,
				Name:         data.CurrencyNames[l][cc],
				FlagEmoji:    cur.FlagEmoji,
			})
		}
	}
	for _, id := range data.TimezoneOrder[l] {
		tz := data.Timezones[id]
		if containsStr(tz.CountryCodes, code) {
			resp.Timezones = append(resp.Timezones, &pb.Timezone{
				Id:           tz.ID,
				Aliases:      tz.Aliases,
				CountryCodes: tz.CountryCodes,
				Name:         data.TimezoneNames[l][id],
			})
		}
	}
	return resp, nil
}

func containsStr(ss []string, want string) bool {
	for _, s := range ss {
		if s == want {
			return true
		}
	}
	return false
}
