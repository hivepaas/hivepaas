package settings

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/permission"
	"github.com/hivepaas/hivepaas/hivepaas_app/repository"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/appservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/clustersecretservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/clusterservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/fileservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/projectservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/settingservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/userservice"
)

type BaseUC struct {
	DB *database.DB

	FileRepo          repository.FileRepo
	SharedSettingRepo repository.SharedSettingRepo
	SettingRepo       repository.SettingRepo

	AppService           appservice.Service
	ClusterSecretService clustersecretservice.Service
	ClusterService       clusterservice.Service
	FileService          fileservice.Service
	ProjectService       projectservice.Service
	SettingService       settingservice.Service
	UserService          userservice.Service
	PermissionManager    permission.Manager
}

func New(
	db *database.DB,

	fileRepo repository.FileRepo,
	sharedSettingRepo repository.SharedSettingRepo,
	settingRepo repository.SettingRepo,

	appService appservice.Service,
	clusterSecretService clustersecretservice.Service,
	clusterService clusterservice.Service,
	fileService fileservice.Service,
	projectService projectservice.Service,
	settingService settingservice.Service,
	userService userservice.Service,
	permissionManager permission.Manager,
) *BaseUC {
	return &BaseUC{
		DB: db,

		FileRepo:          fileRepo,
		SharedSettingRepo: sharedSettingRepo,
		SettingRepo:       settingRepo,

		AppService:           appService,
		ClusterSecretService: clusterSecretService,
		ClusterService:       clusterService,
		FileService:          fileService,
		ProjectService:       projectService,
		SettingService:       settingService,
		UserService:          userService,
		PermissionManager:    permissionManager,
	}
}
