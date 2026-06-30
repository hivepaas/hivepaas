package repocheckoutserviceimpl

import (
	"github.com/localpaas/localpaas/localpaas_app/infra/database"
	"github.com/localpaas/localpaas/localpaas_app/repository"
	"github.com/localpaas/localpaas/localpaas_app/service/repocheckoutservice"
)

type service struct {
	db *database.DB

	fileRepo repository.FileRepo
}

func New(
	db *database.DB,

	fileRepo repository.FileRepo,
) repocheckoutservice.Service {
	return &service{
		db: db,

		fileRepo: fileRepo,
	}
}
