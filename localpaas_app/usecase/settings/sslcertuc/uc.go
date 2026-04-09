package sslcertuc

import (
	"github.com/localpaas/localpaas/localpaas_app/base"
	"github.com/localpaas/localpaas/localpaas_app/entity"
	"github.com/localpaas/localpaas/localpaas_app/service/sslservice"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings"
)

const (
	currentSettingType    = base.SettingTypeSSLCert
	currentSettingVersion = entity.CurrentSSLCertVersion
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
