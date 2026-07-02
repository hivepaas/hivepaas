package envvarserviceimpl

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/repository"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/envvarservice"
)

func New(
	settingRepo repository.SettingRepo,
) envvarservice.Service {
	return &service{
		settingRepo: settingRepo,
	}
}

type service struct {
	settingRepo repository.SettingRepo
}
