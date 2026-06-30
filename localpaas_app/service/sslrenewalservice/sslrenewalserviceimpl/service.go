package sslrenewalserviceimpl

import (
	"github.com/localpaas/localpaas/localpaas_app/infra/database"
	"github.com/localpaas/localpaas/localpaas_app/repository"
	"github.com/localpaas/localpaas/localpaas_app/service/notificationservice"
	"github.com/localpaas/localpaas/localpaas_app/service/sslrenewalservice"
	"github.com/localpaas/localpaas/localpaas_app/service/sslservice"
	"github.com/localpaas/localpaas/localpaas_app/service/traefikservice"
)

type service struct {
	db *database.DB

	settingRepo repository.SettingRepo

	notificationService notificationservice.Service
	sslService          sslservice.Service
	traefikService      traefikservice.Service
}

func New(
	db *database.DB,

	settingRepo repository.SettingRepo,

	notificationService notificationservice.Service,
	sslService sslservice.Service,
	traefikService traefikservice.Service,
) sslrenewalservice.Service {
	return &service{
		db: db,

		settingRepo: settingRepo,

		notificationService: notificationService,
		sslService:          sslService,
		traefikService:      traefikService,
	}
}
