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
		httpSettings *entity.Setting) error

	GetProjectNetworkName(project *entity.Project, env string) string
	GetOrCreateProjectNetwork(ctx context.Context, db database.IDB, project *entity.Project, env string) (
		*entity.Setting, *network.Inspect, error)
	ListProjectNetworks(ctx context.Context, db database.IDB, project *entity.Project) (
		[]*entity.Setting, map[string]*network.Summary, error)
	RemoveAllProjectNetworks(ctx context.Context, db database.IDB, project *entity.Project) error
}
