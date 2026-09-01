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
	"github.com/hivepaas/hivepaas/hivepaas_app/service/scopeservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/settingeventservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/settingservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/userservice"
)

type BaseUC struct {
	DB *database.DB

	FileRepo          repository.FileRepo
	SharedSettingRepo repository.SharedSettingRepo
	SettingRepo       repository.SettingRepo
	TagRepo           repository.TagRepo

	AppService           appservice.Service
	ClusterSecretService clustersecretservice.Service
	ClusterService       clusterservice.Service
	FileService          fileservice.Service
	ProjectService       projectservice.Service
	ScopeService         scopeservice.Service
	SettingEventService  settingeventservice.Service
	SettingService       settingservice.Service
	UserService          userservice.Service
	PermissionManager    permission.Manager
}

func New(
	db *database.DB,

	fileRepo repository.FileRepo,
	sharedSettingRepo repository.SharedSettingRepo,
	settingRepo repository.SettingRepo,
	tagRepo repository.TagRepo,

	appService appservice.Service,
	clusterSecretService clustersecretservice.Service,
	clusterService clusterservice.Service,
	fileService fileservice.Service,
	projectService projectservice.Service,
	scopeService scopeservice.Service,
	settingEventService settingeventservice.Service,
	settingService settingservice.Service,
	userService userservice.Service,
	permissionManager permission.Manager,
) *BaseUC {
	return &BaseUC{
		DB: db,

		FileRepo:          fileRepo,
		SharedSettingRepo: sharedSettingRepo,
		SettingRepo:       settingRepo,
		TagRepo:           tagRepo,

		AppService:           appService,
		ClusterSecretService: clusterSecretService,
		ClusterService:       clusterService,
		FileService:          fileService,
		ProjectService:       projectService,
		ScopeService:         scopeService,
		SettingEventService:  settingEventService,
		SettingService:       settingService,
		UserService:          userService,
		PermissionManager:    permissionManager,
	}
}
