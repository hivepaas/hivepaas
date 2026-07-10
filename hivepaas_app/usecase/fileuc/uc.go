package fileuc

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/repository"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/appservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/fileservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/projectservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/userservice"
)

type UC struct {
	db *database.DB

	fileRepo    repository.FileRepo
	settingRepo repository.SettingRepo

	appService     appservice.Service
	fileService    fileservice.Service
	projectService projectservice.Service
	userService    userservice.Service
}

func New(
	db *database.DB,

	fileRepo repository.FileRepo,
	settingRepo repository.SettingRepo,

	appService appservice.Service,
	fileService fileservice.Service,
	projectService projectservice.Service,
	userService userservice.Service,
) *UC {
	return &UC{
		db: db,

		fileRepo:    fileRepo,
		settingRepo: settingRepo,

		appService:     appService,
		fileService:    fileService,
		projectService: projectService,
		userService:    userService,
	}
}
