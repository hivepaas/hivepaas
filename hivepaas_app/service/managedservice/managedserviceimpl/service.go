package managedserviceimpl

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/repository"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/apptemplateservice/templatemodel"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/managedservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/managedservice/managedmodel"
)

type service struct {
	registry    *registry
	settingRepo repository.SettingRepo
}

func New(settingRepo repository.SettingRepo) managedservice.Service {
	return &service{registry: newRegistry(), settingRepo: settingRepo}
}

func (s *service) Engine(name templatemodel.ManagedEngine) (managedmodel.Engine, bool) {
	return s.registry.Engine(name)
}
