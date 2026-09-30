package schedjobexecserviceimpl

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/repository"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/appservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/commandservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/containerexecservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/fileservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/schedjobexecservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/schedjobservice"
)

type service struct {
	fileRepo    repository.FileRepo
	fileService fileservice.Service

	appService           appservice.Service
	commandService       commandservice.Service
	containerExecService containerexecservice.Service
	schedJobService      schedjobservice.Service
}

func New(
	fileRepo repository.FileRepo,
	fileService fileservice.Service,

	appService appservice.Service,
	commandService commandservice.Service,
	containerExecService containerexecservice.Service,
	schedJobService schedjobservice.Service,
) schedjobexecservice.Service {
	return &service{
		fileRepo:    fileRepo,
		fileService: fileService,

		appService:           appService,
		commandService:       commandService,
		containerExecService: containerExecService,
		schedJobService:      schedJobService,
	}
}
