package sysbackupserviceimpl

import (
	"github.com/localpaas/localpaas/localpaas_app/repository"
	"github.com/localpaas/localpaas/localpaas_app/service/sysbackupservice"
)

type service struct {
	fileRepo repository.FileRepo
}

func New(
	fileRepo repository.FileRepo,
) sysbackupservice.Service {
	return &service{
		fileRepo: fileRepo,
	}
}
