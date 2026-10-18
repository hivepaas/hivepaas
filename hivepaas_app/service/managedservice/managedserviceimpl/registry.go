package managedserviceimpl

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/service/apptemplateservice/templatemodel"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/managedservice/managedmodel"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/managedservice/postgres"
)

// registry is every engine this HivePaaS has. A test holds it to
// templatemodel.ManagedEngines, the engines templates may name.
type registry struct {
	engines []managedmodel.Engine
}

func newRegistry() *registry {
	return &registry{engines: []managedmodel.Engine{
		postgres.New(),
	}}
}

func (r *registry) Engine(name templatemodel.ManagedEngine) (managedmodel.Engine, bool) {
	for _, engine := range r.engines {
		if engine.Name() == name {
			return engine, true
		}
	}
	return nil, false
}
