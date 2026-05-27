package schedjobserviceimpl

import (
	"github.com/localpaas/localpaas/localpaas_app/repository"
	"github.com/localpaas/localpaas/localpaas_app/service/envvarservice"
	"github.com/localpaas/localpaas/localpaas_app/service/schedjobservice"
)

func New(
	settingRepo repository.SettingRepo,
	envVarService envvarservice.Service,
) schedjobservice.Service {
	return &service{
		settingRepo:   settingRepo,
		envVarService: envVarService,
	}
}

type service struct {
	settingRepo   repository.SettingRepo
	envVarService envvarservice.Service
}
