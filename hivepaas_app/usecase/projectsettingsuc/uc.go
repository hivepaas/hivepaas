package projectsettingsuc

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/permission"
	"github.com/hivepaas/hivepaas/hivepaas_app/repository"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/projectservice"
)

type UC struct {
	db *database.DB

	projectRepo              repository.ProjectRepo
	projectSharedSettingRepo repository.ProjectSharedSettingRepo
	settingRepo              repository.SettingRepo

	projectService projectservice.Service

	permissionManager permission.Manager
}

func New(
	db *database.DB,

	projectRepo repository.ProjectRepo,
	projectSharedSettingRepo repository.ProjectSharedSettingRepo,
	settingRepo repository.SettingRepo,

	projectService projectservice.Service,

	permissionManager permission.Manager,
) *UC {
	return &UC{
		db: db,

		projectRepo:              projectRepo,
		projectSharedSettingRepo: projectSharedSettingRepo,
		settingRepo:              settingRepo,

		projectService: projectService,

		permissionManager: permissionManager,
	}
}
