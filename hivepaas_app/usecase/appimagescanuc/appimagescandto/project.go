package appimagescandto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
)

type GetProjectImageScanReq struct {
	ProjectID string `json:"-"`
}

func NewGetProjectImageScanReq() *GetProjectImageScanReq {
	return &GetProjectImageScanReq{}
}

func (req *GetProjectImageScanReq) Validate() hperrors.ValidationErrors {
	return hperrors.NewValidationErrors(vld.Validate())
}

type GetProjectImageScanResp struct {
	Meta *basedto.Meta             `json:"meta"`
	Data *ProjectImageScanDataResp `json:"data"`
}

// ProjectImageScanDataResp is a project's Security tab: its scanned apps and
// what their images carry, the worst first - all of them, as a project has a
// screenful - and whether scanning is on, which a project's members cannot read
// in System settings.
type ProjectImageScanDataResp struct {
	Enabled bool              `json:"enabled"`
	Apps    []*ScannedAppResp `json:"apps"`
}
