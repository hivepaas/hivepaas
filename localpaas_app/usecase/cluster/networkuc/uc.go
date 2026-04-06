package networkuc

import (
	"github.com/localpaas/localpaas/localpaas_app/infra/database"
	"github.com/localpaas/localpaas/localpaas_app/repository"
	"github.com/localpaas/localpaas/localpaas_app/service/clusterservice"
	"github.com/localpaas/localpaas/localpaas_app/service/projectservice"
	"github.com/localpaas/localpaas/services/docker"
)

type NetworkUC struct {
	db             *database.DB
	settingRepo    repository.SettingRepo
	clusterService clusterservice.ClusterService
	projectService projectservice.ProjectService
	dockerManager  docker.Manager
}

func NewNetworkUC(
	db *database.DB,
	settingRepo repository.SettingRepo,
	clusterService clusterservice.ClusterService,
	projectService projectservice.ProjectService,
	dockerManager docker.Manager,
) *NetworkUC {
	return &NetworkUC{
		db:             db,
		settingRepo:    settingRepo,
		clusterService: clusterService,
		projectService: projectService,
		dockerManager:  dockerManager,
	}
}
