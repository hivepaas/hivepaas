package networkservice

import (
	"context"

	"github.com/moby/moby/api/types/network"
	"github.com/moby/moby/api/types/swarm"

	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
)

type Service interface {
	GetGlobalRoutingNetworkID(ctx context.Context) (string, error)
	UpdateAppGlobalRoutingNetwork(ctx context.Context, app *entity.App, service *swarm.Service,
		httpSettings *entity.Setting) error

	GetProjectNetworkName(project *entity.Project, env string) string
	GetOrCreateProjectNetwork(ctx context.Context, project *entity.Project, env string) (*network.Inspect, error)
	ListProjectNetworks(ctx context.Context, project *entity.Project) ([]network.Summary, error)
	RemoveProjectNetwork(ctx context.Context, project *entity.Project, env string) error
	RemoveAllProjectNetworks(ctx context.Context, project *entity.Project) error
}
