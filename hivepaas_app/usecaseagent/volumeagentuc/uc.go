package volumeagentuc

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/logging"
	"github.com/hivepaas/hivepaas/services/docker"
)

type UC struct {
	logger        logging.Logger
	dockerManager docker.Manager
}

func New(
	logger logging.Logger,
	dockerManager docker.Manager,
) *UC {
	return &UC{
		logger:        logger,
		dockerManager: dockerManager,
	}
}
