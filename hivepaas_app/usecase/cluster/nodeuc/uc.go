package nodeuc

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/repository"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/clusterservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/hpappservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/licenseservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/schedjobservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/tasks/queue"
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
	taskRepo      repository.TaskRepo
	taskQueue     queue.TaskQueue

	clusterService  clusterservice.Service
	hpAppService    hpappservice.Service
	licenseService  licenseservice.Service
	schedJobService schedjobservice.Service

	*settings.BaseUC
}

func New(
	db *database.DB,
	dockerManager docker.Manager,
	taskRepo repository.TaskRepo,
	taskQueue queue.TaskQueue,

	clusterService clusterservice.Service,
	hpAppService hpappservice.Service,
	licenseService licenseservice.Service,
	schedJobService schedjobservice.Service,

	baseUC *settings.BaseUC,
) *UC {
	return &UC{
		db:            db,
		dockerManager: dockerManager,
		taskRepo:      taskRepo,
		taskQueue:     taskQueue,

		clusterService:  clusterService,
		hpAppService:    hpAppService,
		licenseService:  licenseService,
		schedJobService: schedJobService,

		BaseUC: baseUC,
	}
}
