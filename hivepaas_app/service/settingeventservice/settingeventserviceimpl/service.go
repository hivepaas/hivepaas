package settingeventserviceimpl

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/repository/cacherepository"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/registryauthservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/settingeventservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/settingmountservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/systemeventbusservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/tasks/queue"
)

func New(
	periodicSettingsRepo cacherepository.PeriodicSettingsRepo,

	registryAuthService registryauthservice.Service,
	settingMountService settingmountservice.Service,
	systemEventBus systemeventbusservice.Service,
	taskQueue queue.TaskQueue,
) settingeventservice.Service {
	return &service{
		periodicSettingsRepo: periodicSettingsRepo,

		registryAuthService: registryAuthService,
		settingMountService: settingMountService,
		systemEventBus:      systemEventBus,
		taskQueue:           taskQueue,
	}
}

type service struct {
	periodicSettingsRepo cacherepository.PeriodicSettingsRepo

	// registryAuthService records a renewal for the ECR credentials an edit
	// changes the keys of.
	registryAuthService registryauthservice.Service
	settingMountService settingmountservice.Service
	systemEventBus      systemeventbusservice.Service
	taskQueue           queue.TaskQueue
}
