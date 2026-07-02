package projectuc

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/permission"
	"github.com/hivepaas/hivepaas/hivepaas_app/repository"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/appservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/projectservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/userservice"
)

type UC struct {
	db *database.DB

	binObjectRepo repository.BinObjectRepo
	projectRepo   repository.ProjectRepo

	appService     appservice.Service
	projectService projectservice.Service
	userService    userservice.Service

	permissionManager permission.Manager
}

func New(
	db *database.DB,

	binObjectRepo repository.BinObjectRepo,
	projectRepo repository.ProjectRepo,

	appService appservice.Service,
	projectService projectservice.Service,
	userService userservice.Service,

	permissionManager permission.Manager,
) *UC {
	return &UC{
		db: db,

		binObjectRepo: binObjectRepo,
		projectRepo:   projectRepo,

		appService:     appService,
		projectService: projectService,
		userService:    userService,

		permissionManager: permissionManager,
	}
}
