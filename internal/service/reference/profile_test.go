package reference

import (
	"context"
	"testing"

	pb "github.com/servekit/api/gen/go/reference/v1"

	"github.com/servekit/reference-service/internal/data"
)

func codes(gs []*pb.RegionGroup) []string {
	out := make([]string, 0, len(gs))
	for _, g := range gs {
		out = append(out, g.GetName())
	}
	return out
}

func TestGetCountryProfileCN(t *testing.T) {
	s := New()
	resp, err := s.GetCountryProfile(context.Background(), &pb.GetCountryProfileRequest{RegionCode: "CN"})
	if err != nil {
		t.Fatal(err)
	}
	if resp.GetCountry().GetName() != "中国" || resp.GetCountry().GetDialCode() != "+86" {
		t.Fatalf("country = %+v", resp.GetCountry())
	}
	if got := resp.GetCountry().GetLanguageTags(); len(got) != 1 || got[0] != "zh" {
		t.Fatalf("CN country.language_tags = %v, want [zh]", got)
	}
	if got := codes(resp.GetRegionGroups()); len(got) != 2 || got[0] != "亚洲" || got[1] != "东亚" {
		t.Fatalf("CN chain = %v, want [亚洲 东亚]", got)
	}
	// CNY first; CLDR also lists CNH (offshore renminbi) as valid for CN.
	if len(resp.GetCurrencies()) < 1 || resp.GetCurrencies()[0].GetCode() != "CNY" {
		t.Fatalf("CN currencies = %+v", resp.GetCurrencies())
	}
	// zone1970 lists two CN zones: Asia/Shanghai and Asia/Urumqi.
	if len(resp.GetTimezones()) != 2 || resp.GetTimezones()[0].GetId() != "Asia/Shanghai" || resp.GetTimezones()[1].GetId() != "Asia/Urumqi" {
		t.Fatalf("CN timezones = %+v", resp.GetTimezones())
	}
	if len(resp.GetLanguages()) == 0 || resp.GetLanguages()[0].GetTag() != "zh" {
		t.Fatalf("CN languages = %+v", resp.GetLanguages())
	}
	if resp.GetDataVersion() != data.Version {
		t.Fatalf("data_version mismatch")
	}
}

func TestGetCountryProfileUSLocaleEN(t *testing.T) {
	s := New()
	resp, err := s.GetCountryProfile(context.Background(), &pb.GetCountryProfileRequest{RegionCode: "US", Locale: "en"})
	if err != nil {
		t.Fatal(err)
	}
	// M49 is three levels for the US: 019 Americas -> 003 North America
	// (CLDR naming) -> 021 Northern America.
	if got := codes(resp.GetRegionGroups()); len(got) != 3 || got[0] != "Americas" || got[2] != "Northern America" {
		t.Fatalf("US chain = %v, want [Americas North America Northern America]", got)
	}
	if len(resp.GetTimezones()) < 5 {
		t.Fatalf("US should list many zones, got %d", len(resp.GetTimezones()))
	}
	if len(resp.GetCurrencies()) != 1 || resp.GetCurrencies()[0].GetCode() != "USD" {
		t.Fatalf("US currencies = %+v", resp.GetCurrencies())
	}
	if len(resp.GetLanguages()) != 1 || resp.GetLanguages()[0].GetTag() != "en" {
		t.Fatalf("US languages = %+v", resp.GetLanguages())
	}
}

func TestGetCountryProfileCHFourOfficial(t *testing.T) {
	s := New()
	resp, _ := s.GetCountryProfile(context.Background(), &pb.GetCountryProfileRequest{RegionCode: "CH"})
	tags := make([]string, 0, len(resp.GetLanguages()))
	for _, l := range resp.GetLanguages() {
		tags = append(tags, l.GetTag())
	}
	// de/fr/it official (gsw is de facto but outside the 639-1 directory set).
	if len(tags) != 3 || tags[0] != "de" || tags[1] != "fr" || tags[2] != "it" {
		t.Fatalf("CH languages = %v, want [de fr it]", tags)
	}
}

func TestGetCountryProfileNotFound(t *testing.T) {
	s := New()
	_, err := s.GetCountryProfile(context.Background(), &pb.GetCountryProfileRequest{RegionCode: "ZZ"})
	if err == nil {
		t.Fatal("expected not-found error")
	}
}

// TestGetCountryProfileAC locks in the curated enrichment for the
// exceptionally-reserved territories: alpha-3 from the ISO reserved list
// (ASC), region folded into Saint Helena's UN chain (Africa -> Western
// Africa), language from the population fallback (en), timezone mirrored
// from SH (Africa/Abidjan — tzdb links Atlantic/St_Helena there).
func TestGetCountryProfileAC(t *testing.T) {
	s := New()
	resp, err := s.GetCountryProfile(context.Background(), &pb.GetCountryProfileRequest{RegionCode: "AC"})
	if err != nil {
		t.Fatal(err)
	}
	if resp.GetCountry().GetAlpha_3() != "ASC" {
		t.Fatalf("AC alpha3 = %q, want ASC", resp.GetCountry().GetAlpha_3())
	}
	if got := codes(resp.GetRegionGroups()); len(got) != 2 || got[0] != "非洲" || got[1] != "西非" {
		t.Fatalf("AC chain = %v, want [非洲 西非]", got)
	}
	if len(resp.GetLanguages()) != 1 || resp.GetLanguages()[0].GetTag() != "en" {
		t.Fatalf("AC languages = %+v, want [en]", resp.GetLanguages())
	}
	if len(resp.GetCurrencies()) != 1 || resp.GetCurrencies()[0].GetCode() != "SHP" {
		t.Fatalf("AC currencies = %+v, want [SHP]", resp.GetCurrencies())
	}
	if len(resp.GetTimezones()) != 1 || resp.GetTimezones()[0].GetId() != "Africa/Abidjan" {
		t.Fatalf("AC timezones = %+v, want [Africa/Abidjan]", resp.GetTimezones())
	}

	// Defaults resolve through the same enrichment; XK is mirrored into
	// Europe/Belgrade's membership (not defaults-only), so its profile
	// lists the zone too.
	d, err := s.GetCountryDefaults(context.Background(), &pb.GetCountryDefaultsRequest{RegionCode: "XK"})
	if err != nil {
		t.Fatal(err)
	}
	if d.GetTimezoneId() != "Europe/Belgrade" {
		t.Fatalf("XK default timezone = %q, want Europe/Belgrade", d.GetTimezoneId())
	}
	pk, _ := s.GetCountryProfile(context.Background(), &pb.GetCountryProfileRequest{RegionCode: "XK"})
	if len(pk.GetTimezones()) != 1 || pk.GetTimezones()[0].GetId() != "Europe/Belgrade" {
		t.Fatalf("XK profile timezones = %+v, want [Europe/Belgrade]", pk.GetTimezones())
	}
}
