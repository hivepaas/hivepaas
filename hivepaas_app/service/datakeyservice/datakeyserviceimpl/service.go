package datakeyserviceimpl

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/repository"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/datakeyservice"
)

type service struct {
	encryptionKeyRepo repository.EncryptionKeyRepo
}

func New(
	encryptionKeyRepo repository.EncryptionKeyRepo,
) datakeyservice.Service {
	return &service{
		encryptionKeyRepo: encryptionKeyRepo,
	}
}
