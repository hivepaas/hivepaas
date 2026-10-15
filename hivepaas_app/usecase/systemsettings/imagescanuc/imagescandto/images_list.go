package imagescandto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/appimagescanuc/appimagescandto"
)

type ListImageScanAppsReq struct {
	Paging basedto.Paging `json:"-"`
}

func NewListImageScanAppsReq() *ListImageScanAppsReq {
	return &ListImageScanAppsReq{}
}

// PagingReq is where a handler parses the page asked for.
func (req *ListImageScanAppsReq) PagingReq() *basedto.Paging {
	return &req.Paging
}

func (req *ListImageScanAppsReq) Validate() hperrors.ValidationErrors {
	return hperrors.NewValidationErrors(vld.Validate())
}

type ListImageScanAppsResp struct {
	Meta *basedto.ListMeta                 `json:"meta"`
	Data []*appimagescandto.ScannedAppResp `json:"data"`
}
