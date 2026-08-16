package apphttpservice

import (
	"github.com/moby/moby/api/types/swarm"

	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
)

type ApplyAppHttpReq struct {
	App                       *entity.App
	HttpSettings              *entity.AppHttpSettings
	RefObjects                *entity.RefObjects
	Service                   *swarm.Service
	SkipApplyingSslCerts      bool
	ForceRecreateSslCertFiles bool
	SkipApplyingNetworks      bool
	SkipUpdatingService       bool
}

type ApplyAppHttpResp struct {
	Service *swarm.Service
}
