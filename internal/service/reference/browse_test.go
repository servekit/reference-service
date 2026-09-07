package reference

import (
	"context"
	"testing"

	pb "github.com/servekit/api/gen/go/reference/v1"

	"github.com/servekit/reference-service/internal/data"
)

func TestListCountriesByRegion(t *testing.T) {
	s := New()
	ctx := context.Background()

	ea, err := s.ListCountriesByRegion(ctx, &pb.ListCountriesByRegionRequest{RegionCode: "030"})
	if err != nil {
		t.Fatal(err)
	}
	want := map[string]bool{"CN": true, "HK": true, "JP": true, "KP": true, "KR": true, "MN": true, "MO": true, "TW": true}
	if len(ea.GetCountries()) != len(want) {
		t.Fatalf("East Asia = %d countries, want %d", len(ea.GetCountries()), len(want))
	}
	for _, c := range ea.GetCountries() {
		if !want[c.GetCode()] {
			t.Fatalf("unexpected %s in East Asia", c.GetCode())
		}
	}

	asia, err := s.ListCountriesByRegion(ctx, &pb.ListCountriesByRegionRequest{RegionCode: "142", Locale: "en"})
	if err != nil {
		t.Fatal(err)
	}
	if len(asia.GetCountries()) < 40 {
		t.Fatalf("Asia recursive = %d countries, want 40+", len(asia.GetCountries()))
	}
	seenCN := false
	for _, c := range asia.GetCountries() {
		if c.GetCode() == "CN" {
			seenCN = true
			if c.GetExampleNumber() == "" {
				t.Fatalf("CN example number missing")
			}
			if got := c.GetLanguageTags(); len(got) == 0 || got[0] != "zh" {
				t.Fatalf("CN language_tags = %v, want zh first", got)
			}
		}
	}
	if !seenCN {
		t.Fatalf("CN not under Asia recursively")
	}

	if _, err := s.ListCountriesByRegion(ctx, &pb.ListCountriesByRegionRequest{RegionCode: "999"}); err == nil {
		t.Fatal("expected region-group-not-found error")
	}
}

func TestGetCountryDefaults(t *testing.T) {
	s := New()
	ctx := context.Background()

	cn, err := s.GetCountryDefaults(ctx, &pb.GetCountryDefaultsRequest{CountryCode: "CN"})
	if err != nil {
		t.Fatal(err)
	}
	if cn.GetTimezoneId() != "Asia/Shanghai" || cn.GetTimezoneName() != "上海" ||
		cn.GetCurrencyCode() != "CNY" || cn.GetCurrencyName() != "人民币" || cn.GetCurrencySymbol() != "¥" ||
		cn.GetLanguageTag() != "zh" || cn.GetDialCode() != "+86" || cn.GetExampleNumber() == "" {
		t.Fatalf("CN defaults = %+v", cn)
	}

	us, _ := s.GetCountryDefaults(ctx, &pb.GetCountryDefaultsRequest{CountryCode: "US", Locale: "en"})
	if us.GetTimezoneId() != "America/New_York" || us.GetCurrencyCode() != "USD" ||
		us.GetLanguageTag() != "en" || us.GetTimezoneName() != "New York" {
		t.Fatalf("US defaults = %+v", us)
	}

	// UA: zone1970's first-row rule lands on the RU,UA Simferopol row; the
	// CLDR exception (spelled with the historic alias Europe/Kiev) must be
	// alias-normalized to Europe/Kyiv.
	ua, _ := s.GetCountryDefaults(ctx, &pb.GetCountryDefaultsRequest{CountryCode: "UA"})
	if ua.GetTimezoneId() != "Europe/Kyiv" {
		t.Fatalf("UA default timezone = %q, want Europe/Kyiv", ua.GetTimezoneId())
	}
	// TW: CLDR's zh_Hant (underscore) must normalize to zh-Hant.
	tw, _ := s.GetCountryDefaults(ctx, &pb.GetCountryDefaultsRequest{CountryCode: "TW"})
	if tw.GetLanguageTag() != "zh-Hant" {
		t.Fatalf("TW default language = %q, want zh-Hant", tw.GetLanguageTag())
	}

	if _, err := s.GetCountryDefaults(ctx, &pb.GetCountryDefaultsRequest{CountryCode: "ZZ"}); err == nil {
		t.Fatal("expected country-not-found error")
	}
}

func TestGetDataInfo(t *testing.T) {
	s := New()
	info, err := s.GetDataInfo(context.Background(), &pb.GetDataInfoRequest{})
	if err != nil {
		t.Fatal(err)
	}
	if info.GetDataVersion() != data.Version ||
		len(info.GetLocales()) != len(data.Locales) ||
		info.GetCountryCount() != int32(len(data.Countries)) ||
		info.GetTimezoneCount() != int32(len(data.Timezones)) ||
		info.GetLanguageCount() != int32(len(data.Languages)) ||
		info.GetCurrencyCount() != int32(len(data.Currencies)) ||
		info.GetRegionGroupCount() != int32(len(data.RegionGroups)) {
		t.Fatalf("info mismatch: %+v", info)
	}
}
