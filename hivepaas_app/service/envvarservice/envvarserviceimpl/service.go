package envvarserviceimpl

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/repository"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/envvarservice"
)

func New(
	appRepo repository.AppRepo,
	settingRepo repository.SettingRepo,
) envvarservice.Service {
	return &service{
		appRepo:     appRepo,
		settingRepo: settingRepo,
	}
}

type service struct {
	appRepo     repository.AppRepo
	settingRepo repository.SettingRepo
}
