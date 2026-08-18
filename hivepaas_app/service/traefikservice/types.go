package traefikservice

import (
	"github.com/moby/moby/api/types/swarm"

	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
)

type ApplyAppConfigReq struct {
	App          *entity.App
	Service      *swarm.Service
	HttpSettings *entity.AppHttpSettings
	RefObjects   *entity.RefObjects
}

type ApplyAppConfigResp struct {
	Service *swarm.Service
}

type RemoveAppConfigReq struct {
	App     *entity.App
	Service *swarm.Service
}

type RemoveAppConfigResp struct {
}
