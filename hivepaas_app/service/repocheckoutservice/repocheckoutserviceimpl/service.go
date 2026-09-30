package repocheckoutserviceimpl

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/repository"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/fileservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/repocheckoutservice"
)

type service struct {
	db *database.DB

	fileRepo    repository.FileRepo
	fileService fileservice.Service
}

func New(
	db *database.DB,

	fileRepo repository.FileRepo,
	fileService fileservice.Service,
) repocheckoutservice.Service {
	return &service{
		db: db,

		fileRepo:    fileRepo,
		fileService: fileService,
	}
}
