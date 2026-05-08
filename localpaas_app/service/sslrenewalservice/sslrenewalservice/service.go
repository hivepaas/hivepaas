package sslrenewalserviceimpl

import (
	"github.com/localpaas/localpaas/localpaas_app/infra/database"
	"github.com/localpaas/localpaas/localpaas_app/infra/logging"
	"github.com/localpaas/localpaas/localpaas_app/repository"
	"github.com/localpaas/localpaas/localpaas_app/service/notificationservice"
	"github.com/localpaas/localpaas/localpaas_app/service/sslrenewalservice"
	"github.com/localpaas/localpaas/localpaas_app/service/sslservice"
	"github.com/localpaas/localpaas/localpaas_app/service/traefikservice"
)

type service struct {
	logger              logging.Logger
	db                  *database.DB
	settingRepo         repository.SettingRepo
	sslService          sslservice.Service
	notificationService notificationservice.Service
	traefikService      traefikservice.Service
}

func New(
	logger logging.Logger,
	db *database.DB,
	settingRepo repository.SettingRepo,
	sslService sslservice.Service,
	notificationService notificationservice.Service,
	traefikService traefikservice.Service,
) sslrenewalservice.Service {
	return &service{
		logger:              logger,
		db:                  db,
		settingRepo:         settingRepo,
		sslService:          sslService,
		notificationService: notificationService,
		traefikService:      traefikService,
	}
}
