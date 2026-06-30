package sslcertuc

import (
	"github.com/localpaas/localpaas/localpaas_app/base"
	"github.com/localpaas/localpaas/localpaas_app/entity"
	"github.com/localpaas/localpaas/localpaas_app/service/domainservice"
	"github.com/localpaas/localpaas/localpaas_app/service/sslservice"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings"
)

const (
	currentSettingType    = base.SettingTypeSSLCert
	currentSettingVersion = entity.CurrentSSLCertVersion
)

type UC struct {
	domainService domainservice.Service
	sslService    sslservice.Service

	*settings.BaseUC
}

func New(
	domainService domainservice.Service,
	sslService sslservice.Service,

	baseUC *settings.BaseUC,
) *UC {
	return &UC{
		domainService: domainService,
		sslService:    sslService,

		BaseUC: baseUC,
	}
}
