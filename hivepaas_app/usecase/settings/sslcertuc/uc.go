package sslcertuc

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/domainservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/sslservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
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
