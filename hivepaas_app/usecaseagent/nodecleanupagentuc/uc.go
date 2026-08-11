package nodecleanupagentuc

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/logging"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/clustercleanupservice"
	"github.com/hivepaas/hivepaas/services/docker"
)

type UC struct {
	logger        logging.Logger
	db            *database.DB
	dockerManager docker.Manager

	clusterCleanupService clustercleanupservice.Service
}

func New(
	logger logging.Logger,
	db *database.DB,
	dockerManager docker.Manager,

	clusterCleanupService clustercleanupservice.Service,
) *UC {
	return &UC{
		logger:        logger,
		db:            db,
		dockerManager: dockerManager,

		clusterCleanupService: clusterCleanupService,
	}
}
