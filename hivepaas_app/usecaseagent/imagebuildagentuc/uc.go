package imagebuildagentuc

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
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

	// tempBaseDir is where the sources sent for a build are unpacked, by day.
	tempBaseDir string
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

		tempBaseDir: base.BaseTempDirDefault,
	}
}

// WithTempBaseDir makes the agent unpack sources under dir.
func (uc *UC) WithTempBaseDir(dir string) *UC {
	uc.tempBaseDir = dir
	return uc
}
