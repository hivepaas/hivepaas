package envvarserviceimpl

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/repository"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/envvarservice"
)

func New(
	appRepo repository.AppRepo,
	resLinkRepo repository.ResLinkRepo,
	settingRepo repository.SettingRepo,
) envvarservice.Service {
	return &service{
		appRepo:     appRepo,
		resLinkRepo: resLinkRepo,
		settingRepo: settingRepo,
	}
}

type service struct {
	appRepo     repository.AppRepo
	resLinkRepo repository.ResLinkRepo
	settingRepo repository.SettingRepo
}
