package pkg

import (
	"context"
	"testing"

	pb "github.com/servekit/api/gen/go/reference/v1"

	"github.com/servekit/reference-service/internal/data"
	"github.com/servekit/reference-service/pkg/config"

	"github.com/stretchr/testify/require"
)

// TestModule_ServesAllDomains drives the module-mode Handler end to end:
// NewModule wires config -> service -> handler, then one RPC per domain
// returns the compiled tables through the real call path (no gRPC wire).
func TestModule_ServesAllDomains(t *testing.T) {
	hdl, err := NewModule(&config.Config{})
	require.NoError(t, err)
	t.Cleanup(func() { _ = hdl.Stop() })

	ctx := context.Background()

	countries, err := hdl.ListCountries(ctx, &pb.ListCountriesRequest{})
	require.NoError(t, err)
	require.Len(t, countries.GetCountries(), len(data.Countries))
	require.Equal(t, "AL", countries.GetCountries()[0].GetCode())
	require.Equal(t, data.Version, countries.GetDataVersion())

	ja, err := hdl.ListCountries(ctx, &pb.ListCountriesRequest{Locale: "ja"})
	require.NoError(t, err)
	require.NotEmpty(t, ja.GetCountries())
	require.Equal(t, "中国", firstName(t, ja, "CN"))

	timezones, err := hdl.ListTimezones(ctx, &pb.ListTimezonesRequest{})
	require.NoError(t, err)
	require.Len(t, timezones.GetTimezones(), len(data.Timezones))

	languages, err := hdl.ListLanguages(ctx, &pb.ListLanguagesRequest{})
	require.NoError(t, err)
	require.Len(t, languages.GetLanguages(), len(data.Languages))

	currencies, err := hdl.ListCurrencies(ctx, &pb.ListCurrenciesRequest{})
	require.NoError(t, err)
	require.Len(t, currencies.GetCurrencies(), len(data.Currencies))

	groups, err := hdl.ListRegionGroups(ctx, &pb.ListRegionGroupsRequest{})
	require.NoError(t, err)
	require.Len(t, groups.GetRegionGroups(), len(data.RegionGroups))
	require.Empty(t, groups.GetRegionGroups()[0].GetParentCode())

	parsed, err := hdl.ParsePhone(ctx, &pb.ParsePhoneRequest{Raw: "+8613800138000"})
	require.NoError(t, err)
	require.True(t, parsed.GetIsValid())
	require.Equal(t, "CN", parsed.GetCountryCode())

	resolved, err := hdl.ResolveCodes(ctx, &pb.ResolveCodesRequest{
		CountryCodes: []string{"CN", "ZZ"},
		TimezoneIds:  []string{"PRC"},
	})
	require.NoError(t, err)
	require.Len(t, resolved.GetCountries(), 1)
	require.Equal(t, []string{"ZZ"}, resolved.GetMissingCountries())
	require.Equal(t, "Asia/Shanghai", resolved.GetTimezones()[0].GetId())
}

func firstName(t *testing.T, resp *pb.ListCountriesResponse, code string) string {
	t.Helper()
	for _, c := range resp.GetCountries() {
		if c.GetCode() == code {
			return c.GetName()
		}
	}
	t.Fatalf("country %s not in list", code)
	return ""
}
