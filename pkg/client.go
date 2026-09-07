package pkg

import (
	"context"

	commonv1 "github.com/servekit/api/gen/go/common/v1"
	referencev1 "github.com/servekit/api/gen/go/reference/v1"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/emptypb"
)

// Client is a gRPC client for reference-service shaped like *Handler: it implements
// the generated referencev1.ReferenceServiceServer interface (unary methods
// without grpc.CallOption), so a consumer can hold either backend behind the
// provider-defined Service interface — module mode passes the *Handler, grpc
// mode passes the *Client — with no per-consumer adapter.
//
// The UnimplementedReferenceServiceServer embed satisfies the interface's
// mustEmbed guard; every RPC below shadows it with a real delegation. When a
// new RPC is added to the proto, add its delegation here — until then grpc
// mode returns codes.Unimplemented for it.
type Client struct {
	referencev1.UnimplementedReferenceServiceServer

	conn *grpc.ClientConn
	cli  referencev1.ReferenceServiceClient
}

// Compile-time assertion: *Client and *Handler expose the same interface.
var _ referencev1.ReferenceServiceServer = (*Client)(nil)

// NewClient dials reference-service at addr using insecure credentials by default.
// Pass additional DialOptions (e.g., credentials) to override.
func NewClient(addr string, opts ...grpc.DialOption) (*Client, error) {
	dialOpts := append([]grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}, opts...)

	conn, err := grpc.NewClient(addr, dialOpts...)
	if err != nil {
		return nil, err
	}

	return &Client{
		conn: conn,
		cli:  referencev1.NewReferenceServiceClient(conn),
	}, nil
}

// Close releases the underlying gRPC connection.
func (c *Client) Close() error { return c.conn.Close() }

// Ping delegates to the remote reference-service.
func (c *Client) Ping(ctx context.Context, in *emptypb.Empty) (*commonv1.Pong, error) {
	return c.cli.Ping(ctx, in)
}

// ListCountries delegates to the remote reference-service.
func (c *Client) ListCountries(ctx context.Context, in *referencev1.ListCountriesRequest) (*referencev1.ListCountriesResponse, error) {
	return c.cli.ListCountries(ctx, in)
}

// GetCountries delegates to the remote reference-service.
func (c *Client) GetCountries(ctx context.Context, in *referencev1.GetCountriesRequest) (*referencev1.GetCountriesResponse, error) {
	return c.cli.GetCountries(ctx, in)
}

// ListTimezones delegates to the remote reference-service.
func (c *Client) ListTimezones(ctx context.Context, in *referencev1.ListTimezonesRequest) (*referencev1.ListTimezonesResponse, error) {
	return c.cli.ListTimezones(ctx, in)
}

// ListLanguages delegates to the remote reference-service.
func (c *Client) ListLanguages(ctx context.Context, in *referencev1.ListLanguagesRequest) (*referencev1.ListLanguagesResponse, error) {
	return c.cli.ListLanguages(ctx, in)
}

// ListCurrencies delegates to the remote reference-service.
func (c *Client) ListCurrencies(ctx context.Context, in *referencev1.ListCurrenciesRequest) (*referencev1.ListCurrenciesResponse, error) {
	return c.cli.ListCurrencies(ctx, in)
}

// ListRegionGroups delegates to the remote reference-service.
func (c *Client) ListRegionGroups(ctx context.Context, in *referencev1.ListRegionGroupsRequest) (*referencev1.ListRegionGroupsResponse, error) {
	return c.cli.ListRegionGroups(ctx, in)
}

// ParsePhone delegates to the remote reference-service.
func (c *Client) ParsePhone(ctx context.Context, in *referencev1.ParsePhoneRequest) (*referencev1.ParsePhoneResponse, error) {
	return c.cli.ParsePhone(ctx, in)
}

// ResolveCodes delegates to the remote reference-service.
func (c *Client) ResolveCodes(ctx context.Context, in *referencev1.ResolveCodesRequest) (*referencev1.ResolveCodesResponse, error) {
	return c.cli.ResolveCodes(ctx, in)
}

// GetCountryProfile delegates to the remote reference-service.
func (c *Client) GetCountryProfile(ctx context.Context, in *referencev1.GetCountryProfileRequest) (*referencev1.GetCountryProfileResponse, error) {
	return c.cli.GetCountryProfile(ctx, in)
}

// ListCountriesByRegion delegates to the remote reference-service.
func (c *Client) ListCountriesByRegion(ctx context.Context, in *referencev1.ListCountriesByRegionRequest) (*referencev1.ListCountriesByRegionResponse, error) {
	return c.cli.ListCountriesByRegion(ctx, in)
}

// GetCountryDefaults delegates to the remote reference-service.
func (c *Client) GetCountryDefaults(ctx context.Context, in *referencev1.GetCountryDefaultsRequest) (*referencev1.GetCountryDefaultsResponse, error) {
	return c.cli.GetCountryDefaults(ctx, in)
}

// GetDataInfo delegates to the remote reference-service.
func (c *Client) GetDataInfo(ctx context.Context, in *referencev1.GetDataInfoRequest) (*referencev1.GetDataInfoResponse, error) {
	return c.cli.GetDataInfo(ctx, in)
}
