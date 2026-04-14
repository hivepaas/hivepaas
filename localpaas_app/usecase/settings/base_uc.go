package settings

import (
	"github.com/localpaas/localpaas/localpaas_app/infra/database"
	"github.com/localpaas/localpaas/localpaas_app/repository"
	"github.com/localpaas/localpaas/localpaas_app/service/appservice"
	"github.com/localpaas/localpaas/localpaas_app/service/fileservice"
	"github.com/localpaas/localpaas/localpaas_app/service/projectservice"
	"github.com/localpaas/localpaas/localpaas_app/service/settingservice"
	"github.com/localpaas/localpaas/localpaas_app/service/userservice"
)

type BaseUC struct {
	DB                       *database.DB
	SettingRepo              repository.SettingRepo
	ProjectService           projectservice.Service
	AppService               appservice.Service
	UserService              userservice.Service
	ProjectSharedSettingRepo repository.ProjectSharedSettingRepo
	SettingService           settingservice.Service
	FileService              fileservice.Service
}

func New(
	db *database.DB,
	settingRepo repository.SettingRepo,
	projectService projectservice.Service,
	appService appservice.Service,
	userService userservice.Service,
	projectSharedSettingRepo repository.ProjectSharedSettingRepo,
	settingService settingservice.Service,
	fileService fileservice.Service,
) *BaseUC {
	return &BaseUC{
		DB:                       db,
		SettingRepo:              settingRepo,
		ProjectService:           projectService,
		AppService:               appService,
		UserService:              userService,
		ProjectSharedSettingRepo: projectSharedSettingRepo,
		SettingService:           settingService,
		FileService:              fileService,
	}
}
