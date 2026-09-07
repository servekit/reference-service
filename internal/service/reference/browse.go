// browse.go — hierarchy navigation and country defaults: the "user picked
// something, now what" helpers. ListCountriesByRegion walks the M49 subtree
// (continent -> sub-regions -> countries); GetCountryDefaults produces the
// auto-fill set (primary zone, first currency, top official language, dial
// code, example number); GetDataInfo reports snapshot metadata.

package reference

import (
	"context"

	pb "github.com/servekit/api/gen/go/reference/v1"

	"github.com/servekit/reference-service/internal/data"
	"github.com/servekit/reference-service/pkg/xcodes"
)

// ListCountriesByRegion returns every country under a region group,
// recursively, in the request locale's country collation order.
func (*Service) ListCountriesByRegion(_ context.Context, req *pb.ListCountriesByRegionRequest) (*pb.ListCountriesByRegionResponse, error) {
	group := req.GetGroupCode()
	if _, ok := data.RegionGroups[group]; !ok {
		return nil, xcodes.ErrRegionGroupNotFound.Wrapf(nil, "region group %s", group)
	}
	l := resolveLocale(req.GetLocale())

	// Subtree: the group plus every group descending from it. The walk is
	// bounded by the group count — generated data is acyclic (tested), the
	// bound only keeps a future data regression from hanging the RPC.
	inSubtree := map[string]bool{group: true}
	for g, rg := range data.RegionGroups {
		for i, cur := 0, rg.ParentCode; cur != "" && i <= len(data.RegionGroups); i, cur = i+1, data.RegionGroups[cur].ParentCode {
			if cur == group {
				inSubtree[g] = true
				break
			}
		}
	}
	countrySet := map[string]bool{}
	for g := range inSubtree {
		for _, cc := range data.RegionGroups[g].RegionCodes {
			countrySet[cc] = true
		}
	}

	out := make([]*pb.Country, 0, len(countrySet))
	for _, cc := range data.CountryOrder[l] {
		if !countrySet[cc] {
			continue
		}
		out = append(out, countryRow(l, cc, data.Countries[cc]))
	}
	return &pb.ListCountriesByRegionResponse{Countries: out, DataVersion: data.Version}, nil
}

// GetCountryDefaults returns the auto-fill set for one country.
func (*Service) GetCountryDefaults(_ context.Context, req *pb.GetCountryDefaultsRequest) (*pb.GetCountryDefaultsResponse, error) {
	code := req.GetRegionCode()
	c, ok := data.Countries[code]
	if !ok {
		return nil, xcodes.ErrCountryNotFound.Wrapf(nil, "country %s", code)
	}
	l := resolveLocale(req.GetLocale())

	resp := &pb.GetCountryDefaultsResponse{
		DialCode:      c.DialCode,
		ExampleNumber: c.ExampleNumber,
		DataVersion:   data.Version,
	}

	// Primary timezone; fall back to the first zone covering the country
	// in collation order (defense — every zoned country has a primary).
	if tzID := data.PrimaryZones[code]; tzID != "" {
		resp.TimezoneId = tzID
		resp.TimezoneName = data.TimezoneNames[l][tzID]
	} else {
		for _, id := range data.TimezoneOrder[l] {
			if containsStr(data.Timezones[id].RegionCodes, code) {
				resp.TimezoneId = id
				resp.TimezoneName = data.TimezoneNames[l][id]
				break
			}
		}
	}

	// Default currency: the first currently-valid currency in the locale's
	// order (deterministic; CNY sorts before CNH for zh locales).
	for _, cur := range data.CurrencyOrder[l] {
		if containsStr(data.Currencies[cur].RegionCodes, code) {
			resp.CurrencyCode = cur
			resp.CurrencyName = data.CurrencyNames[l][cur]
			resp.CurrencySymbol = data.Currencies[cur].Symbol
			break
		}
	}

	// Default language: the most-spoken official one.
	if langs := data.CountryLanguages[code]; len(langs) > 0 {
		resp.LanguageTag = langs[0]
		resp.LanguageName = data.LanguageNames[l][langs[0]]
	}
	return resp, nil
}

// GetDataInfo reports the compiled snapshot's metadata.
func (*Service) GetDataInfo(_ context.Context, _ *pb.GetDataInfoRequest) (*pb.GetDataInfoResponse, error) {
	return &pb.GetDataInfoResponse{
		DataVersion:      data.Version,
		Locales:          data.Locales,
		RegionCount:      int32(len(data.Countries)),
		TimezoneCount:    int32(len(data.Timezones)),
		LanguageCount:    int32(len(data.Languages)),
		CurrencyCount:    int32(len(data.Currencies)),
		RegionGroupCount: int32(len(data.RegionGroups)),
	}, nil
}
