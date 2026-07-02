package settings

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/repository"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/appservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/clusterservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/fileservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/projectservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/settingservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/userservice"
)

type BaseUC struct {
	DB                       *database.DB
	SettingRepo              repository.SettingRepo
	FileRepo                 repository.FileRepo
	ProjectService           projectservice.Service
	AppService               appservice.Service
	UserService              userservice.Service
	ProjectSharedSettingRepo repository.ProjectSharedSettingRepo
	SettingService           settingservice.Service
	FileService              fileservice.Service
	ClusterService           clusterservice.Service
}

func New(
	db *database.DB,
	settingRepo repository.SettingRepo,
	fileRepo repository.FileRepo,
	projectService projectservice.Service,
	appService appservice.Service,
	userService userservice.Service,
	projectSharedSettingRepo repository.ProjectSharedSettingRepo,
	settingService settingservice.Service,
	fileService fileservice.Service,
	clusterService clusterservice.Service,
) *BaseUC {
	return &BaseUC{
		DB:                       db,
		SettingRepo:              settingRepo,
		FileRepo:                 fileRepo,
		ProjectService:           projectService,
		AppService:               appService,
		UserService:              userService,
		ProjectSharedSettingRepo: projectSharedSettingRepo,
		SettingService:           settingService,
		FileService:              fileService,
		ClusterService:           clusterService,
	}
}
