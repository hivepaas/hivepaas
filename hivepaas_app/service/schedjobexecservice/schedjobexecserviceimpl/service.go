package schedjobexecserviceimpl

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/repository"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/containerexecservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/schedjobexecservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/schedjobservice"
)

type service struct {
	fileRepo repository.FileRepo

	containerExecService containerexecservice.Service
	schedJobService      schedjobservice.Service
}

func New(
	fileRepo repository.FileRepo,

	containerExecService containerexecservice.Service,
	schedJobService schedjobservice.Service,
) schedjobexecservice.Service {
	return &service{
		fileRepo: fileRepo,

		containerExecService: containerExecService,
		schedJobService:      schedJobService,
	}
}
