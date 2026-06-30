package fileuc

import (
	"github.com/localpaas/localpaas/localpaas_app/infra/database"
	"github.com/localpaas/localpaas/localpaas_app/repository"
	"github.com/localpaas/localpaas/localpaas_app/service/appservice"
	"github.com/localpaas/localpaas/localpaas_app/service/fileservice"
	"github.com/localpaas/localpaas/localpaas_app/service/projectservice"
	"github.com/localpaas/localpaas/localpaas_app/service/userservice"
)

type UC struct {
	db *database.DB

	fileRepo repository.FileRepo

	appService     appservice.Service
	fileService    fileservice.Service
	projectService projectservice.Service
	userService    userservice.Service
}

func New(
	db *database.DB,

	fileRepo repository.FileRepo,

	appService appservice.Service,
	fileService fileservice.Service,
	projectService projectservice.Service,
	userService userservice.Service,
) *UC {
	return &UC{
		db: db,

		fileRepo: fileRepo,

		appService:     appService,
		fileService:    fileService,
		projectService: projectService,
		userService:    userService,
	}
}
