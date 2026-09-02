package backupreposerviceimpl

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/repository"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/backupreposervice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/nodeexecservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/settingservice"
	"github.com/hivepaas/hivepaas/services/docker"
)

type service struct {
	db            *database.DB
	dockerManager docker.Manager

	settingRepo repository.SettingRepo
	tagRepo     repository.TagRepo

	nodeExecService nodeexecservice.Service
	settingService  settingservice.Service
}

func New(
	db *database.DB,
	dockerManager docker.Manager,

	settingRepo repository.SettingRepo,
	tagRepo repository.TagRepo,

	nodeExecService nodeexecservice.Service,
	settingService settingservice.Service,
) backupreposervice.Service {
	return &service{
		db:            db,
		dockerManager: dockerManager,

		settingRepo: settingRepo,
		tagRepo:     tagRepo,

		nodeExecService: nodeExecService,
		settingService:  settingService,
	}
}
