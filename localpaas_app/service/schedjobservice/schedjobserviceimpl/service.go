package schedjobserviceimpl

import (
	"github.com/localpaas/localpaas/localpaas_app/service/envvarservice"
	"github.com/localpaas/localpaas/localpaas_app/service/schedjobservice"
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
