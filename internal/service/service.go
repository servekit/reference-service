// Package service contains reference-service business logic.
//
// Layering contract (see golang-service-development skill §2):
//   - This is the SERVICE ROOT. It holds Service struct + New + Start/Stop +
//     one-line facade methods (one per RPC).
//   - Business logic lives in SUBPACKAGES (internal/service/<domain>/). This
//     file does NOT contain CRUD implementations — only delegations.
//   - handler calls service.X; service.X is a one-line facade that calls
//     s.<domain>.X in the subpackage. handler never imports the subpackage.
//   - Service methods take proto types DIRECTLY and return proto types — no
//     intermediate Go structs at any layer.
//   - Resources (db, redis, third-party services) are constructed here from
//     cfg (or injected via option) and passed to subpackage constructors. The
//     subpackages do NOT manage resource lifecycle — this Service does via
//     lifecycle.Manager.
package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	commonv1 "github.com/servekit/api/gen/go/common/v1"
	referencev1 "github.com/servekit/api/gen/go/reference/v1"

	"github.com/servekit/reference-service/internal/jobs"
	"github.com/servekit/reference-service/internal/service/reference"
	"github.com/servekit/reference-service/internal/version"

	"github.com/servekit/reference-service/pkg/config"
	"github.com/servekit/reference-service/pkg/option"

	"github.com/servekit/go-common/cronx"
	"github.com/servekit/go-common/lifecycle"
)

// Service holds reference-service business state.
//
// Resource fields (db, redis, gid) are convenience references kept on the root
// Service — they point at the same instances tracked by mgr and injected into
// subpackages. Each domain lives in its own subpackage field
// (reference *reference.Service); subpackages do NOT reference this struct.
type Service struct {
	cfg *config.Config
	mgr *lifecycle.Manager

	// One field per domain subpackage. The reference domain is pure
	// computation over compiled tables — no resources, no lifecycle.
	reference *reference.Service

	// startedAt is set once in New; Ping returns it for uptime.
	startedAt int64
}

// New constructs a Service from config and functional options.
//
// Resources not injected via options are created from cfg, wrapped as
// lifecycle.Stoppers, and registered with the internal Manager. Stop will
// stop them in reverse order. Injected resources are NOT registered — caller
// owns their lifecycle.
//
// On partial failure (any resolve returns an error), already-registered
// components are stopped via mgr.Stop() before returning the error.
func New(cfg *config.Config, opts ...option.Option) (*Service, error) {
	o := option.Apply(opts...)
	mgr := lifecycle.NewManager()
	_ = o // no resources enabled; o kept for the injection seam

	// jobs.Scheduler owns the cron instance; setupJobs builds it, registers
	// it on mgr, and wires periodic jobs (empty by default — add jobs inside
	// setupJobs as scheduler.AddFunc calls). See architecture.md (jobs.md).
	svc := &Service{
		cfg: cfg,
		mgr: mgr,

		reference: reference.New(),

		startedAt: time.Now().UnixMilli(),
	}

	if err := svc.setupJobs(); err != nil {
		if cerr := mgr.Stop(); cerr != nil {
			err = errors.Join(err, fmt.Errorf("rollback: %w", cerr))
		}
		return nil, err
	}

	return svc, nil
}

// Start starts all owned components concurrently.
func (s *Service) Start() error { return s.mgr.Start() }

// Stop stops all owned components in reverse registration order.
func (s *Service) Stop() error { return s.mgr.Stop() }

// Ping is a health-check RPC, always generated so gRPC clients have a
// uniform liveness probe. Returns only public, non-sensitive info — never
// internal addresses, env, secrets, or dependency topology.
func (s *Service) Ping(_ context.Context) (*commonv1.Pong, error) {
	v := version.Get()
	return &commonv1.Pong{
		Service:   "reference-service",
		Version:   v.Version,
		GitCommit: v.GitCommit,
		GitBranch: v.GitBranch,
		BuildTime: v.BuildTime,
		GoVersion: v.GoVersion,
		Status:    "SERVING",
		Now:       time.Now().UnixMilli(),
		StartedAt: s.startedAt,
	}, nil
}

// Resource resolve helpers (resolveDB / resolveRedis)
// live in helper.go — extracted from this file to keep service.go focused on
// the Service struct, New/Start/Stop/Ping, and the facade delegations.

// setupJobs builds the jobs.Scheduler, registers it on s.mgr, and wires
// periodic jobs. Signature is intentionally receiver-only: future jobs are
// added inside this method as scheduler.AddFunc calls. Timezone default lives
// in config.CronConfig's default tag. Module-mode embedders may pass a
// config without a cron section — no cron configured means no scheduler.
func (s *Service) setupJobs() error {
	if s.cfg == nil || s.cfg.Cron == nil {
		return nil
	}
	scheduler, err := jobs.New(&jobs.Deps{
		Config: &cronx.Config{
			Timezone:      s.cfg.Cron.Timezone,
			OverlapPolicy: "skip",
		},
	})
	if err != nil {
		return fmt.Errorf("init jobs: %w", err)
	}
	s.mgr.Add("jobs", scheduler)
	return nil
}

// --- facade methods (one per RPC, delegate to subpackage) ---

// ListCountries delegates to the reference subpackage.
func (s *Service) ListCountries(ctx context.Context, req *referencev1.ListCountriesRequest) (*referencev1.ListCountriesResponse, error) {
	return s.reference.ListCountries(ctx, req)
}

// ListTimezones delegates to the reference subpackage.
func (s *Service) ListTimezones(ctx context.Context, req *referencev1.ListTimezonesRequest) (*referencev1.ListTimezonesResponse, error) {
	return s.reference.ListTimezones(ctx, req)
}

// ListLanguages delegates to the reference subpackage.
func (s *Service) ListLanguages(ctx context.Context, req *referencev1.ListLanguagesRequest) (*referencev1.ListLanguagesResponse, error) {
	return s.reference.ListLanguages(ctx, req)
}

// ListCurrencies delegates to the reference subpackage.
func (s *Service) ListCurrencies(ctx context.Context, req *referencev1.ListCurrenciesRequest) (*referencev1.ListCurrenciesResponse, error) {
	return s.reference.ListCurrencies(ctx, req)
}

// ListRegionGroups delegates to the reference subpackage.
func (s *Service) ListRegionGroups(ctx context.Context, req *referencev1.ListRegionGroupsRequest) (*referencev1.ListRegionGroupsResponse, error) {
	return s.reference.ListRegionGroups(ctx, req)
}

// ParsePhone delegates to the reference subpackage.
func (s *Service) ParsePhone(ctx context.Context, req *referencev1.ParsePhoneRequest) (*referencev1.ParsePhoneResponse, error) {
	return s.reference.ParsePhone(ctx, req)
}

// ResolveCodes delegates to the reference subpackage.
func (s *Service) ResolveCodes(ctx context.Context, req *referencev1.ResolveCodesRequest) (*referencev1.ResolveCodesResponse, error) {
	return s.reference.ResolveCodes(ctx, req)
}
