package reslinkserviceimpl

import (
	"github.com/localpaas/localpaas/localpaas_app/repository"
	"github.com/localpaas/localpaas/localpaas_app/service/reslinkservice"
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
