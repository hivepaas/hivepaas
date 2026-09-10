package projectenvsettingsuc

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/permission"
	"github.com/hivepaas/hivepaas/hivepaas_app/repository"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/auditservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/envvarservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/projectservice"
)

type UC struct {
	db *database.DB

	projectEnvRepo    repository.ProjectEnvRepo
	projectRepo       repository.ProjectRepo
	sharedSettingRepo repository.SharedSettingRepo
	settingRepo       repository.SettingRepo

	auditService   auditservice.Service
	envVarService  envvarservice.Service
	projectService projectservice.Service

	permissionManager permission.Manager
}

func New(
	db *database.DB,

	projectEnvRepo repository.ProjectEnvRepo,
	projectRepo repository.ProjectRepo,
	sharedSettingRepo repository.SharedSettingRepo,
	settingRepo repository.SettingRepo,

	auditService auditservice.Service,
	envVarService envvarservice.Service,
	projectService projectservice.Service,

	permissionManager permission.Manager,
) *UC {
	return &UC{
		db: db,

		projectEnvRepo:    projectEnvRepo,
		projectRepo:       projectRepo,
		sharedSettingRepo: sharedSettingRepo,
		settingRepo:       settingRepo,

		auditService:   auditService,
		envVarService:  envVarService,
		projectService: projectService,

		permissionManager: permissionManager,
	}
}
