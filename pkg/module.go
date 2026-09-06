package pkg

import (
	referencev1 "github.com/servekit/api/gen/go/reference/v1"

	"github.com/servekit/reference-service/internal/service"
	"github.com/servekit/reference-service/pkg/config"
	"github.com/servekit/reference-service/pkg/handler"
	"github.com/servekit/reference-service/pkg/option"
)

// Handler is the in-process entry point. Callers invoke proto-typed RPC
// methods directly on it — no serialization, no network. This IS the public
// capability surface of reference-service when embedded as a module.
//
// Aliased to *handler.Handler so external code references it as
// referencepkg.Handler without importing internal packages.
type Handler = handler.Handler

// Compile-time assertion: *Handler satisfies the gRPC server interface.
var _ referencev1.ReferenceServiceServer = (*Handler)(nil)

// NewModule constructs an in-process reference service for embedding.
//
// Returns only the Handler — Handler IS the public capability and ALSO
// satisfies signalx.Service (Start/Stop), so module users manage lifecycle
// via the same object they call RPC methods on:
//
//	hdl, err := referencepkg.NewModule(cfg, option.WithDB(parentDB))
//	if err != nil { panic(err) }
//	if err := hdl.Start(); err != nil { panic(err) }   // background goroutines (cron, etc.)
//	defer hdl.Stop()                                    // closes owned resources
//	reference, err := hdl.GetReference(ctx, &referencev1.GetReferenceRequest{Id: 1})
//
// Resources injected via option.WithDB / WithGIDHandler are NOT owned by the
// service — parent process keeps ownership and is responsible for cleanup.
// Only resources the service creates from cfg are tracked by the internal
// lifecycle.Manager and stopped on Stop.
func NewModule(cfg *config.Config, opts ...option.Option) (*Handler, error) {
	svc, err := service.New(cfg, opts...)
	if err != nil {
		return nil, err
	}
	return handler.New(svc), nil
}
