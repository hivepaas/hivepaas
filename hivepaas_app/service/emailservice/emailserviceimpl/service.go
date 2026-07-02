package emailserviceimpl

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/repository"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/emailservice"
)

func New(
	settingRepo repository.SettingRepo,
) emailservice.Service {
	return &service{
		settingRepo: settingRepo,
	}
}

type service struct {
	settingRepo repository.SettingRepo
}
