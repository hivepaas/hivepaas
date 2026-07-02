package domainserviceimpl

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/repository"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/domainservice"
)

func New(
	resLinkRepo repository.ResLinkRepo,
	settingRepo repository.SettingRepo,
) domainservice.Service {
	return &service{
		resLinkRepo: resLinkRepo,
		settingRepo: settingRepo,
	}
}

type service struct {
	resLinkRepo repository.ResLinkRepo
	settingRepo repository.SettingRepo
}
