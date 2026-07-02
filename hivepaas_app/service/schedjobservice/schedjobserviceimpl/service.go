package schedjobserviceimpl

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/service/envvarservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/schedjobservice"
)

func New(
	envVarService envvarservice.Service,
) schedjobservice.Service {
	return &service{
		envVarService: envVarService,
	}
}

type service struct {
	envVarService envvarservice.Service
}
