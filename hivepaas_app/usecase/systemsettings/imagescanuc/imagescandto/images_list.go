package imagescandto

import (
	"time"

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
	Meta *basedto.ListMeta   `json:"meta"`
	Data []*ImageScanAppResp `json:"data"`
}

// ImageScanAppResp is an app and what its images carry, in counts. The apps
// come the worst first: by fixable CRITICAL, then fixable HIGH.
type ImageScanAppResp struct {
	App        *basedto.NamedObjectResp     `json:"app"`
	Project    *basedto.NamedObjectResp     `json:"project"`
	ProjectEnv *basedto.NamedObjectResp     `json:"projectEnv"`
	ScannedAt  time.Time                    `json:"scannedAt"`
	Error      string                       `json:"error,omitempty"`
	Summary    *appimagescandto.SummaryResp `json:"summary"`
	Images     []*appimagescandto.ImageResp `json:"images"`
}
