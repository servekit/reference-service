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
	if first.GetRegionCode() != "AL" || first.GetName() != "阿尔巴尼亚" || first.GetFlagEmoji() == "" {
		t.Fatalf("first zh-Hans entry = %+v, want AL/阿尔巴尼亚 with flag", first)
	}
	if got := first.GetLanguageTags(); len(got) != 1 || got[0] != "sq" {
		t.Fatalf("AL language_tags = %v, want [sq]", got)
	}
	// Territories without official-status entries fall back to population
	// ranking in the generator — AC gets [en] (CLDR: 99% English, no flag).
	for _, c := range resp.GetCountries() {
		if c.GetRegionCode() == "AC" {
			if got := c.GetLanguageTags(); len(got) != 1 || got[0] != "en" {
				t.Fatalf("AC language_tags = %v, want [en]", got)
			}
			if c.GetAlpha_3() != "ASC" {
				t.Fatalf("AC alpha3 = %q, want ASC (UPU-reserved overlay)", c.GetAlpha_3())
			}
			break
		}
	}
	// CLDR spells the script suffix with an underscore (zh_Hant); the
	// generator normalizes — TW must carry its official Chinese.
	for _, c := range resp.GetCountries() {
		if c.GetRegionCode() == "TW" {
			if got := c.GetLanguageTags(); len(got) != 1 || got[0] != "zh-Hant" {
				t.Fatalf("TW language_tags = %v, want [zh-Hant]", got)
			}
			break
		}
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
	if got := resp.GetCountries()[0]; got.GetRegionCode() != "AF" || got.GetName() != "Afghanistan" {
		t.Fatalf("en first entry = %s/%s", got.GetRegionCode(), got.GetName())
	}
	// Same list in ja still returns every country.
	respJa, _ := s.ListCountries(context.Background(), &pb.ListCountriesRequest{Locale: "ja"})
	if len(respJa.GetCountries()) != len(data.Countries) {
		t.Fatalf("ja list incomplete: %d", len(respJa.GetCountries()))
	}
}

func TestGetCountries(t *testing.T) {
	s := New()
	resp, err := s.GetCountries(context.Background(), &pb.GetCountriesRequest{
		RegionCodes: []string{"US", "AC", "ZZ", "CN"},
	})
	if err != nil {
		t.Fatal(err)
	}
	// Rows come back in request order (not locale collation)...
	if len(resp.GetCountries()) != 3 ||
		resp.GetCountries()[0].GetRegionCode() != "US" ||
		resp.GetCountries()[1].GetRegionCode() != "AC" ||
		resp.GetCountries()[2].GetRegionCode() != "CN" {
		t.Fatalf("countries not in request order: %+v", resp.GetCountries())
	}
	// ...carrying the full directory row: enriched alpha-3 + languages.
	if got := resp.GetCountries()[1]; got.GetAlpha_3() != "ASC" || len(got.GetLanguageTags()) != 1 || got.GetLanguageTags()[0] != "en" {
		t.Fatalf("AC row = %+v", got)
	}
	// Unknown codes are reported, never an error.
	if len(resp.GetMissingRegions()) != 1 || resp.GetMissingRegions()[0] != "ZZ" {
		t.Fatalf("missing countries: %v", resp.GetMissingRegions())
	}
	if resp.GetDataVersion() != data.Version {
		t.Fatalf("data_version = %q, want %q", resp.GetDataVersion(), data.Version)
	}

	// Empty request is a valid no-op (nothing requested, nothing missing).
	empty, err := s.GetCountries(context.Background(), &pb.GetCountriesRequest{})
	if err != nil || len(empty.GetCountries()) != 0 || len(empty.GetMissingRegions()) != 0 {
		t.Fatalf("empty request: %+v err=%v", empty, err)
	}

	// Locale resolution applies like every other read.
	en, _ := s.GetCountries(context.Background(), &pb.GetCountriesRequest{RegionCodes: []string{"CN"}, Locale: "en"})
	if en.GetCountries()[0].GetName() != "China" {
		t.Fatalf("en name = %q", en.GetCountries()[0].GetName())
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
		RegionCodes:   []string{"CN", "ZZ"},
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
	// ResolveCodes rows carry the same fields as ListCountries (language_tags).
	if got := resp.GetCountries()[0].GetLanguageTags(); len(got) != 1 || got[0] != "zh" {
		t.Fatalf("CN language_tags = %v, want [zh]", got)
	}
	if len(resp.GetMissingRegions()) != 1 || resp.GetMissingRegions()[0] != "ZZ" {
		t.Fatalf("missing countries: %v", resp.GetMissingRegions())
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
