package fileserviceimpl

import (
	"github.com/localpaas/localpaas/localpaas_app/repository"
	"github.com/localpaas/localpaas/localpaas_app/service/fileservice"
	"github.com/localpaas/localpaas/localpaas_app/service/settingservice"
)

func New(
	fileRepo repository.FileRepo,
	settingRepo repository.SettingRepo,
	settingService settingservice.Service,
) fileservice.Service {
	return &service{
		fileRepo:       fileRepo,
		settingRepo:    settingRepo,
		settingService: settingService,
	}
}

type service struct {
	fileRepo       repository.FileRepo
	settingRepo    repository.SettingRepo
	settingService settingservice.Service
}
