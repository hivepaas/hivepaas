package schedjobexecserviceimpl

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/repository"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/appservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/containerexecservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/schedjobexecservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/schedjobservice"
)

type service struct {
	fileRepo repository.FileRepo

	appService           appservice.Service
	containerExecService containerexecservice.Service
	schedJobService      schedjobservice.Service
}

func New(
	fileRepo repository.FileRepo,

	appService appservice.Service,
	containerExecService containerexecservice.Service,
	schedJobService schedjobservice.Service,
) schedjobexecservice.Service {
	return &service{
		fileRepo: fileRepo,

		appService:           appService,
		containerExecService: containerExecService,
		schedJobService:      schedJobService,
	}
}
