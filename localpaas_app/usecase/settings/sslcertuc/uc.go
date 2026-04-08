package sslcertuc

import (
	"github.com/localpaas/localpaas/localpaas_app/service/sslservice"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings"
)

type UC struct {
	*settings.BaseUC
	sslService sslservice.Service
}

func New(
	baseUC *settings.BaseUC,
	sslService sslservice.Service,
) *UC {
	return &UC{
		BaseUC:     baseUC,
		sslService: sslService,
	}
}
