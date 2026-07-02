package reslinkserviceimpl

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/repository"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/reslinkservice"
)

func New(
	resLinkRepo repository.ResLinkRepo,
) reslinkservice.Service {
	return &service{
		resLinkRepo: resLinkRepo,
	}
}

type service struct {
	resLinkRepo repository.ResLinkRepo
}
