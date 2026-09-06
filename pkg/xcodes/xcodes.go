// Package xcodes defines reference-service error codes.
//
// Only codes actually used by the codebase are defined — add new ones when
// business code requires them (YAGNI). Defined locally via xerr.New so the
// package is self-contained; errors.Is still works across packages because
// xerr.Error.Is compares by reason string.
package xcodes

import "github.com/servekit/go-common/xerr"

// ErrCountryNotFound indicates no country matches the requested alpha-2.
var ErrCountryNotFound = xerr.New("COUNTRY_NOT_FOUND", xerr.CategoryNotFound, 404, "country not found")
