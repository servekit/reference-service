// ParsePhone: libphonenumber rules as a deterministic pure
// function (the metadata ships with the nyaruka dependency). Bad input
// never fails the RPC; is_valid=false + error_reason carries the reason and
// inferable fields are still filled.

package reference

import (
	"context"
	"strconv"
	"strings"

	"github.com/nyaruka/phonenumbers"

	pb "github.com/servekit/api/gen/go/reference/v1"

	"github.com/servekit/reference-service/internal/data"
)

// ParsePhone applies libphonenumber rules; bad input reports is_valid=false.
func (*Service) ParsePhone(_ context.Context, req *pb.ParsePhoneRequest) (*pb.ParsePhoneResponse, error) {
	num, err := phonenumbers.Parse(strings.TrimSpace(req.GetRaw()), req.GetDefaultRegion())
	if err != nil {
		//nolint:nilerr // a parse failure is a valid response (is_valid=false), not an RPC error
		return &pb.ParsePhoneResponse{IsValid: false, ErrorReason: err.Error()}, nil
	}
	resp := &pb.ParsePhoneResponse{
		IsValid:                phonenumbers.IsValidNumber(num),
		E164:                   phonenumbers.Format(num, phonenumbers.E164),
		NationalNumber:         strconv.FormatUint(num.GetNationalNumber(), 10),
		FormattedInternational: phonenumbers.Format(num, phonenumbers.INTERNATIONAL),
		Type:                   phoneTypeToProto(phonenumbers.GetNumberType(num)),
	}
	if region := phonenumbers.GetRegionCodeForNumber(num); region != "" {
		resp.CountryCode = region
		if c, ok := data.Countries[region]; ok {
			resp.DialCode = c.DialCode
		}
	}
	if !resp.IsValid && resp.ErrorReason == "" {
		resp.ErrorReason = "invalid number"
	}
	return resp, nil
}

// phoneTypeToProto maps libphonenumber's PhoneNumberType onto the proto
// enum; names are kept identical so the mapping is mechanical.
func phoneTypeToProto(t phonenumbers.PhoneNumberType) pb.PhoneType {
	switch t {
	case phonenumbers.MOBILE:
		return pb.PhoneType_PHONE_TYPE_MOBILE
	case phonenumbers.FIXED_LINE:
		return pb.PhoneType_PHONE_TYPE_FIXED_LINE
	case phonenumbers.FIXED_LINE_OR_MOBILE:
		return pb.PhoneType_PHONE_TYPE_FIXED_LINE_OR_MOBILE
	case phonenumbers.TOLL_FREE:
		return pb.PhoneType_PHONE_TYPE_TOLL_FREE
	case phonenumbers.PREMIUM_RATE:
		return pb.PhoneType_PHONE_TYPE_PREMIUM_RATE
	case phonenumbers.SHARED_COST:
		return pb.PhoneType_PHONE_TYPE_SHARED_COST
	case phonenumbers.VOIP:
		return pb.PhoneType_PHONE_TYPE_VOIP
	case phonenumbers.PERSONAL_NUMBER:
		return pb.PhoneType_PHONE_TYPE_PERSONAL_NUMBER
	case phonenumbers.PAGER:
		return pb.PhoneType_PHONE_TYPE_PAGER
	case phonenumbers.UAN:
		return pb.PhoneType_PHONE_TYPE_UAN
	case phonenumbers.VOICEMAIL:
		return pb.PhoneType_PHONE_TYPE_VOICE_MAIL
	default:
		return pb.PhoneType_PHONE_TYPE_UNKNOWN
	}
}
