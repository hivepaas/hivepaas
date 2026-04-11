package projectsettingsuc

import (
	"github.com/localpaas/localpaas/localpaas_app/infra/database"
	"github.com/localpaas/localpaas/localpaas_app/permission"
	"github.com/localpaas/localpaas/localpaas_app/repository"
	"github.com/localpaas/localpaas/localpaas_app/service/networkservice"
	"github.com/localpaas/localpaas/localpaas_app/service/projectservice"
	"github.com/localpaas/localpaas/localpaas_app/service/userservice"
	"github.com/localpaas/localpaas/services/docker"
)

type UC struct {
	db                       *database.DB
	projectRepo              repository.ProjectRepo
	settingRepo              repository.SettingRepo
	projectSharedSettingRepo repository.ProjectSharedSettingRepo
	permissionManager        permission.Manager
	userService              userservice.Service
	projectService           projectservice.Service
	networkService           networkservice.Service
	dockerManager            docker.Manager
}

func New(
	db *database.DB,
	projectRepo repository.ProjectRepo,
	settingRepo repository.SettingRepo,
	projectSharedSettingRepo repository.ProjectSharedSettingRepo,
	permissionManager permission.Manager,
	userService userservice.Service,
	projectService projectservice.Service,
	networkService networkservice.Service,
	dockerManager docker.Manager,
) *UC {
	return &UC{
		db:                       db,
		projectRepo:              projectRepo,
		settingRepo:              settingRepo,
		projectSharedSettingRepo: projectSharedSettingRepo,
		permissionManager:        permissionManager,
		userService:              userService,
		projectService:           projectService,
		networkService:           networkService,
		dockerManager:            dockerManager,
	}
}
