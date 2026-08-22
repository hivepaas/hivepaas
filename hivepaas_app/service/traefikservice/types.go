package traefikservice

import (
	"github.com/moby/moby/api/types/swarm"

	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
)

// APP HTTP CONFIG

type ApplyAppConfigReq struct {
	App          *entity.App
	Service      *swarm.Service // if nil, it will be reloaded automatically
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

// TRUSTED IPS IN FORWARDED HEADER

type ApplyTrustedIPsReq struct {
	Service                     *swarm.Service // if nil, it will be reloaded automatically
	TrustedIPs                  []string
	SkipUpdatingServiceInDocker bool
}

type ApplyTrustedIPsResp struct {
	Service *swarm.Service
}

// OPEN TCP/UDP PORTS

type OpenPortReq struct {
	Service                     *swarm.Service // if nil, it will be reloaded automatically
	OpenPorts                   []string       // format: 9000/tcp, 10000/udp
	ClosePorts                  []string
	SkipUpdatingServiceInDocker bool
}

type OpenPortResp struct {
	Service *swarm.Service
}
