package service

// This file holds the resource resolve helpers used by service.New. They were
// extracted from service.go to keep that file focused on the Service struct,
// New/Start/Stop/Ping, and the RPC facade delegations.
//
// Every dependency follows the platform's inject-or-build contract: an
// injected one (option.With…) is used as-is with the caller owning its
// lifecycle; otherwise it is built from cfg via the provider's Connect
// (dbx/redisx for resources, the dependency's pkg.Connect for services),
// which also registers its lifecycle with the Manager.
