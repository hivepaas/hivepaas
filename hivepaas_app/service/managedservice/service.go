// Package managedservice is the engines that look after managed templates' apps.
package managedservice

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/service/apptemplateservice/templatemodel"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/managedservice/managedmodel"
)

type Service interface {
	// Engine is the engine of that name, and false when this HivePaaS has none.
	Engine(name templatemodel.ManagedEngine) (managedmodel.Engine, bool)
}
