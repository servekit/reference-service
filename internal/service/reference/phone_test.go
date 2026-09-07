package reference

import (
	"context"
	"testing"

	pb "github.com/servekit/api/gen/go/reference/v1"
	"github.com/servekit/reference-service/internal/data"
)

func TestParsePhoneGolden(t *testing.T) {
	s := New()
	ctx := context.Background()

	got, err := s.ParsePhone(ctx, &pb.ParsePhoneRequest{Raw: "+8613800138000"})
	if err != nil {
		t.Fatal(err)
	}
	if !got.GetIsValid() || got.GetCountryCode() != "CN" || got.GetDialCode() != "+86" ||
		got.GetE164() != "+8613800138000" || got.GetNationalNumber() != "13800138000" ||
		got.GetType() != pb.PhoneType_PHONE_TYPE_MOBILE {
		t.Fatalf("CN mobile: %+v", got)
	}
	if got.GetFormattedInternational() != "+86 138 0013 8000" {
		t.Logf("CN international format = %q (informational)", got.GetFormattedInternational())
	}

	got2, err := s.ParsePhone(ctx, &pb.ParsePhoneRequest{Raw: "13800138000", DefaultRegion: "CN"})
	if err != nil {
		t.Fatal(err)
	}
	if !got2.GetIsValid() || got2.GetE164() != "+8613800138000" {
		t.Fatalf("local-format CN: %+v", got2)
	}

	bad, err := s.ParsePhone(ctx, &pb.ParsePhoneRequest{Raw: "not-a-phone"})
	if err != nil {
		t.Fatal(err)
	}
	if bad.GetIsValid() || bad.GetErrorReason() == "" {
		t.Fatalf("invalid input must report reason: %+v", bad)
	}
}

func TestParsePhoneUSLine(t *testing.T) {
	s := New()
	got, _ := s.ParsePhone(context.Background(), &pb.ParsePhoneRequest{Raw: "+1 650 253 0000"})
	if !got.GetIsValid() || got.GetCountryCode() != "US" || got.GetDialCode() != "+1" {
		t.Fatalf("US line: %+v", got)
	}
	switch got.GetType() {
	case pb.PhoneType_PHONE_TYPE_FIXED_LINE,
		pb.PhoneType_PHONE_TYPE_FIXED_LINE_OR_MOBILE,
		pb.PhoneType_PHONE_TYPE_MOBILE:
		// US numbering cannot always distinguish — any of these is honest.
	default:
		t.Fatalf("US line type = %v", got.GetType())
	}
}

// TestExampleNumbersAreMobile guards the placeholder data: the example must
// parse as a MOBILE number (regions that cannot distinguish mobile from
// fixed line by prefix are exempt — their type is FIXED_LINE_OR_MOBILE).
func TestExampleNumbersAreMobile(t *testing.T) {
	s := New()
	ctx := context.Background()
	exempt := map[string]bool{"US": true, "CA": true} // NANP: same prefix space
	checked := 0
	for _, cc := range []string{"CN", "AT", "DE", "RU", "JP", "GB", "FR", "BR", "IN", "AU"} {
		c, ok := data.Countries[cc]
		if !ok || c.ExampleNumber == "" {
			t.Fatalf("%s: no example number", cc)
		}
		if exempt[cc] {
			continue
		}
		p, err := s.ParsePhone(ctx, &pb.ParsePhoneRequest{Raw: c.ExampleNumber})
		if err != nil || !p.GetIsValid() {
			t.Fatalf("%s: example %q does not parse: %+v err=%v", cc, c.ExampleNumber, p, err)
		}
		if p.GetType() != pb.PhoneType_PHONE_TYPE_MOBILE {
			t.Errorf("%s: example %q parses as %v, want MOBILE", cc, c.ExampleNumber, p.GetType())
		}
		checked++
	}
	if checked < 8 {
		t.Fatalf("only %d countries checked", checked)
	}
}
