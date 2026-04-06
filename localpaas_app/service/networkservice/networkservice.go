package networkservice

import (
	"context"

	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/api/types/swarm"

	"github.com/localpaas/localpaas/localpaas_app/entity"
	"github.com/localpaas/localpaas/localpaas_app/repository"
	"github.com/localpaas/localpaas/services/docker"
)

type NetworkService interface {
	CreateProjectNetwork(ctx context.Context, project *entity.Project) (*network.CreateResponse, error)
	RemoveProjectNetwork(ctx context.Context, project *entity.Project) error

	UpdateAppGlobalRoutingNetwork(ctx context.Context, app *entity.App, service *swarm.Service,
		httpSettings *entity.Setting) error
}

func NewNetworkService(
	settingRepo repository.SettingRepo,
	dockerManager docker.Manager,
) NetworkService {
	return &networkService{
		settingRepo:   settingRepo,
		dockerManager: dockerManager,
	}
}

type networkService struct {
	settingRepo   repository.SettingRepo
	dockerManager docker.Manager
}
