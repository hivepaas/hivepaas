package getstartedserviceimpl

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/repository"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/domainservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/getstartedservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/hpappservice"
)

type service struct {
	settingRepo      repository.SettingRepo
	systemStatusRepo repository.SystemStatusRepo
	taskRepo         repository.TaskRepo
	domainService    domainservice.Service
	hpAppService     hpappservice.Service
}

func New(
	settingRepo repository.SettingRepo,
	systemStatusRepo repository.SystemStatusRepo,
	taskRepo repository.TaskRepo,
	domainService domainservice.Service,
	hpAppService hpappservice.Service,
) getstartedservice.Service {
	return &service{
		settingRepo:      settingRepo,
		systemStatusRepo: systemStatusRepo,
		taskRepo:         taskRepo,
		domainService:    domainService,
		hpAppService:     hpAppService,
	}
}
