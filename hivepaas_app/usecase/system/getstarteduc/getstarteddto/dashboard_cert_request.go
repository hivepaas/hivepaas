package getstarteddto

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
)

type RequestDashboardCertReq struct{}

func NewRequestDashboardCertReq() *RequestDashboardCertReq {
	return &RequestDashboardCertReq{}
}

// Validate implements interface basedto.ReqValidator
func (req *RequestDashboardCertReq) Validate() hperrors.ValidationErrors {
	return nil
}

type RequestDashboardCertResp struct {
	Meta *basedto.Meta          `json:"meta"`
	Data *DashboardCertDataResp `json:"data"`
}
