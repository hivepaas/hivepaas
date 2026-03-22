package nginxservice

import (
	"context"

	"github.com/docker/docker/api/types/swarm"

	"github.com/localpaas/localpaas/localpaas_app/entity"
	"github.com/localpaas/localpaas/localpaas_app/repository"
	"github.com/localpaas/localpaas/services/docker"
)

type NginxService interface {
	GetNginxSwarmService(ctx context.Context) (*swarm.Service, error)
	RestartNginxSwarmService(ctx context.Context) error

	ReloadNginxConfig(ctx context.Context, restartServiceOnFailure bool) error
	ResetNginxConfig(ctx context.Context) error

	ApplyAppConfig(ctx context.Context, app *entity.App, data *AppConfigData) error
	RemoveAppConfig(ctx context.Context, app *entity.App) error
}

func NewNginxService(
	settingRepo repository.SettingRepo,
	dockerManager docker.Manager,
) NginxService {
	return &nginxService{
		settingRepo:   settingRepo,
		dockerManager: dockerManager,
	}
}

type nginxService struct {
	settingRepo   repository.SettingRepo
	dockerManager docker.Manager
}
