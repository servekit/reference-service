package reference

import (
	"context"
	"testing"

	pb "github.com/servekit/api/gen/go/reference/v1"
	"github.com/servekit/reference-service/internal/data"
)

func TestResolveLocale(t *testing.T) {
	cases := map[string]string{
		"":        "zh-Hans",
		"zh-TW":   "zh-Hant",
		"zh-Hant": "zh-Hant",
		"pt-BR":   "pt",
		"ja":      "ja",
		"en-US":   "en",
		"ru-RU":   "ru",
		"xx":      "en",
		"!@":      "en",
	}
	for in, want := range cases {
		if got := resolveLocale(in); got != want {
			t.Errorf("resolveLocale(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestListCountriesDefaultLocale(t *testing.T) {
	s := New()
	resp, err := s.ListCountries(context.Background(), &pb.ListCountriesRequest{})
	if err != nil {
		t.Fatal(err)
	}
	if len(resp.GetCountries()) != len(data.Countries) {
		t.Fatalf("got %d countries, want %d", len(resp.GetCountries()), len(data.Countries))
	}
	first := resp.GetCountries()[0]
	if first.GetCode() != "AL" || first.GetName() != "阿尔巴尼亚" || first.GetFlagEmoji() == "" {
		t.Fatalf("first zh-Hans entry = %+v, want AL/阿尔巴尼亚 with flag", first)
	}
	if resp.GetDataVersion() != data.Version {
		t.Fatalf("data_version = %q, want %q", resp.GetDataVersion(), data.Version)
	}
}

func TestListCountriesHonorsLocale(t *testing.T) {
	s := New()
	resp, err := s.ListCountries(context.Background(), &pb.ListCountriesRequest{Locale: "en"})
	if err != nil {
		t.Fatal(err)
	}
	if got := resp.GetCountries()[0]; got.GetCode() != "AF" || got.GetName() != "Afghanistan" {
		t.Fatalf("en first entry = %s/%s", got.GetCode(), got.GetName())
	}
	// Same list in ja still returns every country.
	respJa, _ := s.ListCountries(context.Background(), &pb.ListCountriesRequest{Locale: "ja"})
	if len(respJa.GetCountries()) != len(data.Countries) {
		t.Fatalf("ja list incomplete: %d", len(respJa.GetCountries()))
	}
}

func TestListOtherDomains(t *testing.T) {
	s := New()
	ctx := context.Background()
	if tz, _ := s.ListTimezones(ctx, &pb.ListTimezonesRequest{}); len(tz.GetTimezones()) != len(data.Timezones) || tz.GetDataVersion() == "" {
		t.Fatalf("timezones list wrong: %d", len(tz.GetTimezones()))
	}
	if lg, _ := s.ListLanguages(ctx, &pb.ListLanguagesRequest{}); len(lg.GetLanguages()) != len(data.Languages) {
		t.Fatalf("languages list wrong: %d", len(lg.GetLanguages()))
	}
	if cu, _ := s.ListCurrencies(ctx, &pb.ListCurrenciesRequest{}); len(cu.GetCurrencies()) != len(data.Currencies) {
		t.Fatalf("currencies list wrong: %d", len(cu.GetCurrencies()))
	}
	rg, _ := s.ListRegionGroups(ctx, &pb.ListRegionGroupsRequest{})
	if len(rg.GetRegionGroups()) != len(data.RegionGroups) {
		t.Fatalf("region groups list wrong: %d", len(rg.GetRegionGroups()))
	}
	if rg.GetRegionGroups()[0].GetParentCode() != "" {
		t.Fatalf("first region group not top-level: %+v", rg.GetRegionGroups()[0])
	}
}

func TestResolveCodes(t *testing.T) {
	s := New()
	resp, err := s.ResolveCodes(context.Background(), &pb.ResolveCodesRequest{
		CountryCodes:  []string{"CN", "ZZ"},
		TimezoneIds:   []string{"PRC", "Asia/Shanghai"},
		LanguageTags:  []string{"zh-Hans", "qq"},
		CurrencyCodes: []string{"CNY", "BAD"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(resp.GetCountries()) != 1 || resp.GetCountries()[0].GetName() != "中国" {
		t.Fatalf("countries: %+v", resp.GetCountries())
	}
	if len(resp.GetMissingCountries()) != 1 || resp.GetMissingCountries()[0] != "ZZ" {
		t.Fatalf("missing countries: %v", resp.GetMissingCountries())
	}
	if len(resp.GetTimezones()) != 2 || resp.GetTimezones()[0].GetId() != "Asia/Shanghai" {
		t.Fatalf("timezones not alias-normalized: %+v", resp.GetTimezones())
	}
	if len(resp.GetLanguages()) != 1 || resp.GetLanguages()[0].GetTag() != "zh-Hans" {
		t.Fatalf("languages: %+v", resp.GetLanguages())
	}
	if len(resp.GetMissingLanguages()) != 1 || resp.GetMissingLanguages()[0] != "qq" {
		t.Fatalf("missing languages: %v", resp.GetMissingLanguages())
	}
	if len(resp.GetCurrencies()) != 1 || resp.GetCurrencies()[0].GetCode() != "CNY" {
		t.Fatalf("currencies: %+v", resp.GetCurrencies())
	}
	if len(resp.GetMissingCurrencies()) != 1 || resp.GetMissingCurrencies()[0] != "BAD" {
		t.Fatalf("missing currencies: %v", resp.GetMissingCurrencies())
	}
}

func TestResolveCodesLanguageCaseCanonicalized(t *testing.T) {
	s := New()
	resp, _ := s.ResolveCodes(context.Background(), &pb.ResolveCodesRequest{
		LanguageTags: []string{"ZH-HANS"},
	})
	if len(resp.GetLanguages()) != 1 || resp.GetLanguages()[0].GetTag() != "zh-Hans" {
		t.Fatalf("expected case-canonicalized hit, got %+v missing=%v",
			resp.GetLanguages(), resp.GetMissingLanguages())
	}
}
