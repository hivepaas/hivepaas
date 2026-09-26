package getstarteddto

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/getstartedservice"
)

// DashboardCertDataResp is where the dashboard's certificate stands, for the Get
// started card: todo, obtaining, failed or done.
type DashboardCertDataResp struct {
	Status string `json:"status"`
	// Domain is the name the certificate is for.
	Domain string `json:"domain,omitempty"`
	// Error is why the last attempt failed.
	Error string `json:"error,omitempty"`
}

func TransformDashboardCert(item *getstartedservice.Item) *DashboardCertDataResp {
	return &DashboardCertDataResp{
		Status: string(item.Status),
		Domain: item.Domain,
		Error:  item.Error,
	}
}

type GetDashboardCertReq struct{}

func NewGetDashboardCertReq() *GetDashboardCertReq {
	return &GetDashboardCertReq{}
}

type GetDashboardCertResp struct {
	Meta *basedto.Meta          `json:"meta"`
	Data *DashboardCertDataResp `json:"data"`
}
