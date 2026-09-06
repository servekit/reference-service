package pkg

import (
	"context"

	referencev1 "github.com/servekit/api/gen/go/reference/v1"
	commonv1 "github.com/servekit/api/gen/go/common/v1"

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

