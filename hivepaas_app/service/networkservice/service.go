package networkservice

import (
	"context"

	"github.com/moby/moby/api/types/network"
	"github.com/moby/moby/api/types/swarm"

	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
)

type Service interface {
	GetGlobalRoutingNetworkID(ctx context.Context) (string, error)
	UpdateAppGlobalRoutingNetwork(ctx context.Context, app *entity.App, service *swarm.Service,
		routingSettings *entity.AppRoutingSettings) error

	ListProjectNetworks(ctx context.Context, db database.IDB, project *entity.Project) (
		[]*entity.Setting, map[string]*network.Summary, error)
	RemoveAllProjectNetworks(ctx context.Context, db database.IDB, project *entity.Project) error

	GetProjectNetworkName(project *entity.Project, env string) string
	GetOrCreateProjectNetwork(ctx context.Context, db database.IDB, project *entity.Project, env string) (
		*entity.Setting, *network.Inspect, error)

	ListProjectEnvNetworks(ctx context.Context, db database.IDB, projectEnv *entity.ProjectEnv) (
		[]*entity.Setting, map[string]*network.Summary, error)
	RemoveAllProjectEnvNetworks(ctx context.Context, db database.IDB, projectEnv *entity.ProjectEnv) error

	// Sync networks from Docker to app DB
	SyncNetworks(ctx context.Context, db database.IDB) ([]network.Summary, error)
}
