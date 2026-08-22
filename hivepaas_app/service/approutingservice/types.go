package approutingservice

import (
	"github.com/moby/moby/api/types/swarm"

	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
)

type ApplyAppRoutingReq struct {
	App                       *entity.App
	RoutingSettings           *entity.AppRoutingSettings
	RefObjects                *entity.RefObjects
	Service                   *swarm.Service
	SkipApplyingSslCerts      bool
	ForceRecreateSslCertFiles bool
	SkipApplyingNetworks      bool
	SkipUpdatingService       bool
}

type ApplyAppRoutingResp struct {
	Service *swarm.Service
}
