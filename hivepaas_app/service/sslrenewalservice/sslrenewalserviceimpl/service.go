package sslrenewalserviceimpl

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/repository"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/notificationservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/settingservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/sslrenewalservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/sslservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/traefikservice"
)

type service struct {
	db *database.DB

	settingRepo repository.SettingRepo

	notificationService notificationservice.Service
	settingService      settingservice.Service
	sslService          sslservice.Service
	traefikService      traefikservice.Service
}

func New(
	db *database.DB,

	settingRepo repository.SettingRepo,

	notificationService notificationservice.Service,
	settingService settingservice.Service,
	sslService sslservice.Service,
	traefikService traefikservice.Service,
) sslrenewalservice.Service {
	return &service{
		db: db,

		settingRepo: settingRepo,

		notificationService: notificationService,
		settingService:      settingService,
		sslService:          sslService,
		traefikService:      traefikService,
	}
}
