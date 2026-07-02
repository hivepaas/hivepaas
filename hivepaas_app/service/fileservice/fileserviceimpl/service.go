package fileserviceimpl

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/repository"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/fileservice"
)

func New(
	fileRepo repository.FileRepo,
) fileservice.Service {
	return &service{
		fileRepo: fileRepo,
	}
}

type service struct {
	fileRepo repository.FileRepo
}
