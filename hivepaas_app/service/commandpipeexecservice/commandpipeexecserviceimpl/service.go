package commandpipeexecserviceimpl

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/repository"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/appservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/commandpipeexecservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/commandservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/containerexecservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/schedjobservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/settingservice"
	"github.com/hivepaas/hivepaas/services/docker"
)

type service struct {
	fileRepo    repository.FileRepo
	settingRepo repository.SettingRepo

	appService           appservice.Service
	commandService       commandservice.Service
	containerExecService containerexecservice.Service
	schedJobService      schedjobservice.Service
	settingService       settingservice.Service

	dockerManager docker.Manager
}

func New(
	fileRepo repository.FileRepo,
	settingRepo repository.SettingRepo,

	appService appservice.Service,
	commandService commandservice.Service,
	containerExecService containerexecservice.Service,
	schedJobService schedjobservice.Service,
	settingService settingservice.Service,

	dockerManager docker.Manager,
) commandpipeexecservice.Service {
	return &service{
		fileRepo:    fileRepo,
		settingRepo: settingRepo,

		appService:           appService,
		commandService:       commandService,
		containerExecService: containerExecService,
		schedJobService:      schedJobService,
		settingService:       settingService,

		dockerManager: dockerManager,
	}
}
