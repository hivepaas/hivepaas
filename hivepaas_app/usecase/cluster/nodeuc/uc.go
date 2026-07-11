package nodeuc

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/clusterservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/hpappservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
	"github.com/hivepaas/hivepaas/services/docker"
)

const (
	currentSettingType    = base.SettingTypeClusterNode
	currentSettingVersion = entity.CurrentClusterNodeVersion
)

type UC struct {
	db            *database.DB
	dockerManager docker.Manager

	clusterService clusterservice.Service
	hpAppService   hpappservice.Service

	*settings.BaseUC
}

func New(
	db *database.DB,
	dockerManager docker.Manager,

	clusterService clusterservice.Service,
	hpAppService hpappservice.Service,

	baseUC *settings.BaseUC,
) *UC {
	return &UC{
		db:            db,
		dockerManager: dockerManager,

		clusterService: clusterService,
		hpAppService:   hpAppService,

		BaseUC: baseUC,
	}
}
