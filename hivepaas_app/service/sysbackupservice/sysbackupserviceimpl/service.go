package sysbackupserviceimpl

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/repository"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/sysbackupservice"
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
