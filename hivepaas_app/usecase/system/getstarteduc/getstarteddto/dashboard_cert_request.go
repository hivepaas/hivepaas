package getstarteddto

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
)

type RequestDashboardCertReq struct{}

func NewRequestDashboardCertReq() *RequestDashboardCertReq {
	return &RequestDashboardCertReq{}
}

type RequestDashboardCertResp struct {
	Meta *basedto.Meta          `json:"meta"`
	Data *DashboardCertDataResp `json:"data"`
}
