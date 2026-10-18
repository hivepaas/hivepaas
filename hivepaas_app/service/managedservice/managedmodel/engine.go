// Package managedmodel is what every managed engine is to HivePaaS, apart from
// any one of them. See docs/superpowers/specs/2026-10-17-managed-templates-design.md §3.
package managedmodel

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/service/apptemplateservice/templatemodel"
)

// Engine looks after the apps made from the managed templates that name it.
// Each method is added by the task that needs it, so an engine is never asked
// for what it does not yet do.
type Engine interface {
	Name() templatemodel.ManagedEngine
	// APIRange is the engine API versions this engine speaks: the contracts
	// between it and the templates that name it.
	APIRange() (minAPI, maxAPI int)
}
