package projectsettingsuc

import (
	"github.com/localpaas/localpaas/localpaas_app/infra/database"
	"github.com/localpaas/localpaas/localpaas_app/permission"
	"github.com/localpaas/localpaas/localpaas_app/repository"
	"github.com/localpaas/localpaas/localpaas_app/service/projectservice"
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
