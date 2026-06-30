package projectuc

import (
	"github.com/localpaas/localpaas/localpaas_app/infra/database"
	"github.com/localpaas/localpaas/localpaas_app/permission"
	"github.com/localpaas/localpaas/localpaas_app/repository"
	"github.com/localpaas/localpaas/localpaas_app/service/appservice"
	"github.com/localpaas/localpaas/localpaas_app/service/projectservice"
	"github.com/localpaas/localpaas/localpaas_app/service/userservice"
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
