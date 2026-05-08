package notificationserviceimpl

import (
	"github.com/localpaas/localpaas/localpaas_app/repository"
	"github.com/localpaas/localpaas/localpaas_app/service/emailservice"
	"github.com/localpaas/localpaas/localpaas_app/service/notificationservice"
	"github.com/localpaas/localpaas/localpaas_app/service/settingservice"
	"github.com/localpaas/localpaas/localpaas_app/service/userservice"
)

func New(
	settingRepo repository.SettingRepo,
	settingService settingservice.Service,
	userService userservice.Service,
	emailService emailservice.Service,
) notificationservice.Service {
	return &service{
		settingRepo:    settingRepo,
		settingService: settingService,
		userService:    userService,
		emailService:   emailService,
	}
}

type service struct {
	settingRepo    repository.SettingRepo
	settingService settingservice.Service
	userService    userservice.Service
	emailService   emailservice.Service
}
