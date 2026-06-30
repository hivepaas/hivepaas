package fileserviceimpl

import (
	"github.com/localpaas/localpaas/localpaas_app/repository"
	"github.com/localpaas/localpaas/localpaas_app/service/fileservice"
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
