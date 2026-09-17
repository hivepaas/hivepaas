package appprovisionservice

import (
	"context"

	"github.com/moby/moby/api/types/swarm"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
)

// ConfigureFunc fills in an app's configuration. It runs once the app and its
// initial service spec are prepared and before the service is created. The
// settings it returns replace the app's default settings of the same type; what
// it writes into spec is created with the service.
type ConfigureFunc func(ctx context.Context, db database.IDB, app *entity.App, spec *swarm.ServiceSpec) (
	[]*entity.Setting, error)

type ProvisionAppReq struct {
	ProjectID    string
	ProjectEnvID string
	Name         string
	Status       base.AppStatus
	Note         string
	Tags         []string
	// Configure is nil for an empty app.
	Configure ConfigureFunc
}

type ProvisionAppResp struct {
	// App has Project, ProjectEnv and Settings set, and the ServiceID of the
	// service created for it.
	App *entity.App
}
