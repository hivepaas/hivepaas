package commandserviceimpl

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/service/commandservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/envvarservice"
)

func New(
	envVarService envvarservice.Service,
) commandservice.Service {
	return &service{
		envVarService: envVarService,
	}
}

type service struct {
	envVarService envvarservice.Service
}
