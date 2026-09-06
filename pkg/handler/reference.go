// Package handler implements reference.v1.ReferenceServiceServer as a thin shim over
// internal/service. Each method is a one-line delegation — service takes the
// proto request directly (convert at the store boundary, not here).
//
// Handlers hold NO business logic and NO conversion logic. Anything beyond
// `return h.svc.X(ctx, req)` belongs in internal/service.
//
// Handler also implements signalx.Service (Start/Stop) by delegating to the
// underlying Service, so in-process module users manage lifecycle via the same
// object they call RPC methods on.
package handler

import (
	"context"

	commonv1 "github.com/servekit/api/gen/go/common/v1"
	referencev1 "github.com/servekit/api/gen/go/reference/v1"

	"github.com/servekit/reference-service/internal/service"

	"google.golang.org/protobuf/types/known/emptypb"
)

// Handler implements reference.v1.ReferenceServiceServer. It holds no mutable
// state — the embedded *service.Service owns all business state and lifecycle.
type Handler struct {
	referencev1.UnimplementedReferenceServiceServer

	svc *service.Service
}

// New constructs a Handler wrapping svc.
func New(svc *service.Service) *Handler {
	return &Handler{svc: svc}
}

// Compile-time assertion: Handler implements the gRPC server interface.
var _ referencev1.ReferenceServiceServer = (*Handler)(nil)

// Start starts service-internal components (background goroutines for owned
// resources like cron, message consumers, etc.).
func (h *Handler) Start() error { return h.svc.Start() }

// Stop releases resources owned by the service. After Stop, the Handler must
// not be used.
func (h *Handler) Stop() error { return h.svc.Stop() }

// Ping is a health-check RPC, always generated so gRPC clients (and the
// gateway/testkit module embedder) have a uniform liveness probe.
func (h *Handler) Ping(ctx context.Context, _ *emptypb.Empty) (*commonv1.Pong, error) {
	return h.svc.Ping(ctx)
}

// ListCountries delegates to service.ListCountries.
func (h *Handler) ListCountries(ctx context.Context, req *referencev1.ListCountriesRequest) (*referencev1.ListCountriesResponse, error) {
	return h.svc.ListCountries(ctx, req)
}

// ListTimezones delegates to service.ListTimezones.
func (h *Handler) ListTimezones(ctx context.Context, req *referencev1.ListTimezonesRequest) (*referencev1.ListTimezonesResponse, error) {
	return h.svc.ListTimezones(ctx, req)
}

// ListLanguages delegates to service.ListLanguages.
func (h *Handler) ListLanguages(ctx context.Context, req *referencev1.ListLanguagesRequest) (*referencev1.ListLanguagesResponse, error) {
	return h.svc.ListLanguages(ctx, req)
}

// ListCurrencies delegates to service.ListCurrencies.
func (h *Handler) ListCurrencies(ctx context.Context, req *referencev1.ListCurrenciesRequest) (*referencev1.ListCurrenciesResponse, error) {
	return h.svc.ListCurrencies(ctx, req)
}

// ListRegionGroups delegates to service.ListRegionGroups.
func (h *Handler) ListRegionGroups(ctx context.Context, req *referencev1.ListRegionGroupsRequest) (*referencev1.ListRegionGroupsResponse, error) {
	return h.svc.ListRegionGroups(ctx, req)
}

// ParsePhone delegates to service.ParsePhone.
func (h *Handler) ParsePhone(ctx context.Context, req *referencev1.ParsePhoneRequest) (*referencev1.ParsePhoneResponse, error) {
	return h.svc.ParsePhone(ctx, req)
}

// ResolveCodes delegates to service.ResolveCodes.
func (h *Handler) ResolveCodes(ctx context.Context, req *referencev1.ResolveCodesRequest) (*referencev1.ResolveCodesResponse, error) {
	return h.svc.ResolveCodes(ctx, req)
}
