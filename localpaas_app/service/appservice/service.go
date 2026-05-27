package appservice

import (
	"context"

	"github.com/moby/moby/api/types/swarm"
	"github.com/moby/moby/client"

	"github.com/localpaas/localpaas/localpaas_app/base"
	"github.com/localpaas/localpaas/localpaas_app/entity"
	"github.com/localpaas/localpaas/localpaas_app/infra/database"
	"github.com/localpaas/localpaas/localpaas_app/pkg/bunex"
	"github.com/localpaas/localpaas/services/docker"
)

type Service interface {
	LoadApps(ctx context.Context, db database.IDB, projectID string, appIDs []string,
		requireProjectActive, requireAppsActive bool, extraOpts ...bunex.SelectQueryOption) (
		[]*entity.App, error)
	LoadApp(ctx context.Context, db database.IDB, projectID, appID string,
		requireProjectActive, requireAppActive bool, extraOpts ...bunex.SelectQueryOption) (
		*entity.App, error)
	LoadAppByToken(ctx context.Context, db database.IDB, appToken string,
		requireProjectActive, requireAppActive bool, extraOpts ...bunex.SelectQueryOption) (
		*entity.App, error)

	PersistAppData(ctx context.Context, db database.IDB, data *PersistingAppData) error
	DeleteApp(ctx context.Context, app *entity.App) error
	OnAppStatusChanged(ctx context.Context, app *entity.App, oldStatus base.AppStatus) error

	ServiceInspect(ctx context.Context, serviceID string, caching bool) (*swarm.Service, error)
	ServiceUpdate(ctx context.Context, serviceID string, version *swarm.Version, service *swarm.ServiceSpec,
		options ...docker.ServiceUpdateOption) (*client.ServiceUpdateResult, error)

	// Docker secrets
	CreateSwarmSecret(ctx context.Context, db database.IDB, app *entity.App, secret *entity.Secret) error
	UpdateSwarmSecret(ctx context.Context, db database.IDB, app *entity.App, oldSecret, newSecret *entity.Secret) error
	DeleteSwarmSecret(ctx context.Context, db database.IDB, app *entity.App, secret *entity.Secret) error

	// Docker config
	CreateSwarmConfig(ctx context.Context, db database.IDB, app *entity.App, secret *entity.ConfigFile) error
	UpdateSwarmConfig(ctx context.Context, db database.IDB, app *entity.App, oldSecret, newSecret *entity.ConfigFile) error
	DeleteSwarmConfig(ctx context.Context, db database.IDB, app *entity.App, secret *entity.ConfigFile) error
}
