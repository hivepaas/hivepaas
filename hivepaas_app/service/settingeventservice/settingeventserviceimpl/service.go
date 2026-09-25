package settingeventserviceimpl

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/repository/cacherepository"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/settingeventservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/settingmountservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/systemeventbusservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/tasks/queue"
)

func New(
	periodicSettingsRepo cacherepository.PeriodicSettingsRepo,

	settingMountService settingmountservice.Service,
	systemEventBus systemeventbusservice.Service,
	taskQueue queue.TaskQueue,
) settingeventservice.Service {
	return &service{
		periodicSettingsRepo: periodicSettingsRepo,

		settingMountService: settingMountService,
		systemEventBus:      systemEventBus,
		taskQueue:           taskQueue,
	}
}

type service struct {
	periodicSettingsRepo cacherepository.PeriodicSettingsRepo

	settingMountService settingmountservice.Service
	systemEventBus      systemeventbusservice.Service
	taskQueue           queue.TaskQueue
}
