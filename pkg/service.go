package pkg

import (
	referencev1 "github.com/servekit/api/gen/go/reference/v1"
)

// Service is how a consumer holds reference-service regardless of backend: the
// in-process *Handler (module mode) and the gRPC *Client both satisfy it. It
// embeds the generated server interface so the method set tracks the proto
// automatically — no hand-maintained method list here.
type Service interface {
	referencev1.ReferenceServiceServer
}
