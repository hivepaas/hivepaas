package imageuc

import (
	"github.com/localpaas/localpaas/localpaas_app/infra/database"
	"github.com/localpaas/localpaas/localpaas_app/repository"
	"github.com/localpaas/localpaas/localpaas_app/service/clusterservice"
	"github.com/localpaas/localpaas/services/docker"
)

type UC struct {
	db             *database.DB
	settingRepo    repository.SettingRepo
	clusterService clusterservice.Service
	dockerManager  docker.Manager
}

func New(
	db *database.DB,
	settingRepo repository.SettingRepo,
	clusterService clusterservice.Service,
	dockerManager docker.Manager,
) *UC {
	return &UC{
		db:             db,
		settingRepo:    settingRepo,
		clusterService: clusterService,
		dockerManager:  dockerManager,
	}
}
