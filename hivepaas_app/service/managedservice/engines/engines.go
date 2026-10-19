// Package engines is every managed engine this HivePaaS has, by name. The
// managed service looks its engines up here, and so does what that service
// depends on and so cannot ask: the system environment of every app.
package engines

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/service/apptemplateservice/templatemodel"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/managedservice/managedmodel"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/managedservice/postgres"
)

// all is every engine. A test holds it to templatemodel.ManagedEngines, the
// engines templates may name. Engines keep no state, so one of each serves all.
var all = []managedmodel.Engine{
	postgres.New(),
}

// Lookup is the engine of that name, and false when this HivePaaS has none.
func Lookup(name templatemodel.ManagedEngine) (managedmodel.Engine, bool) {
	for _, engine := range all {
		if engine.Name() == name {
			return engine, true
		}
	}
	return nil, false
}
