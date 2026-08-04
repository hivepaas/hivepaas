package projectuc

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/permission"
	"github.com/hivepaas/hivepaas/hivepaas_app/repository"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/appservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/projectservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/userservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/volumeservice"
	"github.com/hivepaas/hivepaas/services/docker"
)

type UC struct {
	db *database.DB

	appRepo       repository.AppRepo
	binObjectRepo repository.BinObjectRepo
	projectRepo   repository.ProjectRepo

	appService     appservice.Service
	projectService projectservice.Service
	userService    userservice.Service
	volumeService  volumeservice.Service

	dockerManager     docker.Manager
	permissionManager permission.Manager
}

func New(
	db *database.DB,

	appRepo repository.AppRepo,
	binObjectRepo repository.BinObjectRepo,
	projectRepo repository.ProjectRepo,

	appService appservice.Service,
	projectService projectservice.Service,
	userService userservice.Service,
	volumeService volumeservice.Service,

	dockerManager docker.Manager,
	permissionManager permission.Manager,
) *UC {
	return &UC{
		db: db,

		appRepo:       appRepo,
		binObjectRepo: binObjectRepo,
		projectRepo:   projectRepo,

		appService:     appService,
		projectService: projectService,
		userService:    userService,
		volumeService:  volumeService,

		dockerManager:     dockerManager,
		permissionManager: permissionManager,
	}
}
