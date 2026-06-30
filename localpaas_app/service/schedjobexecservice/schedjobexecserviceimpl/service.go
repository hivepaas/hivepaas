package schedjobexecserviceimpl

import (
	"github.com/localpaas/localpaas/localpaas_app/repository"
	"github.com/localpaas/localpaas/localpaas_app/service/containerexecservice"
	"github.com/localpaas/localpaas/localpaas_app/service/schedjobexecservice"
	"github.com/localpaas/localpaas/localpaas_app/service/schedjobservice"
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
