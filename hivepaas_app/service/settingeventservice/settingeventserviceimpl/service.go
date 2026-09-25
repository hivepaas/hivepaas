package settingeventserviceimpl

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/repository/cacherepository"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/settingeventservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/settingmountservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/systemeventbusservice"
)

func New(
	periodicSettingsRepo cacherepository.PeriodicSettingsRepo,

	settingMountService settingmountservice.Service,
	systemEventBus systemeventbusservice.Service,
) settingeventservice.Service {
	return &service{
		periodicSettingsRepo: periodicSettingsRepo,

		settingMountService: settingMountService,
		systemEventBus:      systemEventBus,
	}
}

type service struct {
	periodicSettingsRepo cacherepository.PeriodicSettingsRepo

	settingMountService settingmountservice.Service
	systemEventBus      systemeventbusservice.Service
}
