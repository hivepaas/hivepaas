package fileuc

import (
	"github.com/localpaas/localpaas/localpaas_app/infra/database"
	"github.com/localpaas/localpaas/localpaas_app/repository"
	"github.com/localpaas/localpaas/localpaas_app/service/fileservice"
)

type UC struct {
	db          *database.DB
	fileRepo    repository.FileRepo
	fileService fileservice.Service
}

func New(
	db *database.DB,
	fileRepo repository.FileRepo,
	fileService fileservice.Service,
) *UC {
	return &UC{
		db:          db,
		fileRepo:    fileRepo,
		fileService: fileService,
	}
}
