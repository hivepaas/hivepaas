package appservice

import (
	"context"

	"github.com/docker/docker/api/types/swarm"

	"github.com/localpaas/localpaas/localpaas_app/entity"
	"github.com/localpaas/localpaas/localpaas_app/infra/database"
	"github.com/localpaas/localpaas/localpaas_app/pkg/bunex"
	"github.com/localpaas/localpaas/services/docker"
)

type Service interface {
	LoadApp(ctx context.Context, db database.IDB, projectID, appID string,
		requireProjectActive, requireAppActive bool, extraOpts ...bunex.SelectQueryOption) (
		*entity.App, error)
	LoadAppByToken(ctx context.Context, db database.IDB, appToken string,
		requireProjectActive, requireAppActive bool, extraOpts ...bunex.SelectQueryOption) (
		*entity.App, error)

	PersistAppData(ctx context.Context, db database.IDB, data *PersistingAppData) error
	DeleteApp(ctx context.Context, app *entity.App) error

	ServiceInspect(ctx context.Context, serviceID string, caching bool) (*swarm.Service, error)
	ServiceUpdate(ctx context.Context, serviceID string, version *swarm.Version, service *swarm.ServiceSpec,
		options ...docker.ServiceUpdateOption) (*swarm.ServiceUpdateResponse, error)

	CreateDeployment(app *entity.App, deploymentSettings *entity.AppDeploymentSettings) (
		*entity.Deployment, *entity.Task, error)

	CreateSwarmSecret(ctx context.Context, db database.IDB, app *entity.App, secret *entity.Secret) error
	UpdateSwarmSecret(ctx context.Context, db database.IDB, app *entity.App, oldSecret, newSecret *entity.Secret) error
	DeleteSwarmSecret(ctx context.Context, db database.IDB, app *entity.App, secret *entity.Secret) error
}
