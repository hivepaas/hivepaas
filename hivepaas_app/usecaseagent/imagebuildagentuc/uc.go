package imagebuildagentuc

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/logging"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/appservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/imagebuildservice"
	"github.com/hivepaas/hivepaas/services/docker"
)

type UC struct {
	logger        logging.Logger
	db            *database.DB
	dockerManager docker.Manager

	appService        appservice.Service
	imageBuildService imagebuildservice.Service
}

func New(
	logger logging.Logger,
	db *database.DB,
	dockerManager docker.Manager,

	appService appservice.Service,
	imageBuildService imagebuildservice.Service,
) *UC {
	return &UC{
		logger:        logger,
		db:            db,
		dockerManager: dockerManager,

		appService:        appService,
		imageBuildService: imageBuildService,
	}
}
