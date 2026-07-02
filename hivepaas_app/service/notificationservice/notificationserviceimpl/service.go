package notificationserviceimpl

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/repository"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/notificationservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/settingservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/userservice"
)

func New(
	settingRepo repository.SettingRepo,

	settingService settingservice.Service,
	userService userservice.Service,
) notificationservice.Service {
	return &service{
		settingRepo: settingRepo,

		settingService: settingService,
		userService:    userService,
	}
}

type service struct {
	settingRepo repository.SettingRepo

	settingService settingservice.Service
	userService    userservice.Service
}
