package imagescandto

import (
	"time"

	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/imagescanservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/appimagescanuc/appimagescandto"
)

type ListImageScanComponentsReq struct{}

func NewListImageScanComponentsReq() *ListImageScanComponentsReq {
	return &ListImageScanComponentsReq{}
}

func (req *ListImageScanComponentsReq) Validate() hperrors.ValidationErrors {
	return hperrors.NewValidationErrors(vld.Validate())
}

type ListImageScanComponentsResp struct {
	Meta *basedto.Meta                `json:"meta"`
	Data *ImageScanComponentsDataResp `json:"data"`
}

// ImageScanComponentsDataResp is what HivePaaS's own parts run: those that run
// on the installation and were scanned, in HivePaaS's order.
type ImageScanComponentsDataResp struct {
	Components []*ScannedComponentResp `json:"components"`
}

// ComponentRefResp is a part of HivePaaS: its key, which its address names,
// and its name.
type ComponentRefResp struct {
	Key  string `json:"key"`
	Name string `json:"name"`
}

// ScannedComponentResp is a part of HivePaaS and what its images carry, in
// counts, after the global ignores.
type ScannedComponentResp struct {
	Component *ComponentRefResp            `json:"component"`
	ScannedAt time.Time                    `json:"scannedAt"`
	Error     string                       `json:"error,omitempty"`
	Summary   *appimagescandto.SummaryResp `json:"summary"`
	Images    []*appimagescandto.ImageResp `json:"images"`
}

type GetImageScanComponentReq struct {
	Component string `json:"-"`
}

func NewGetImageScanComponentReq() *GetImageScanComponentReq {
	return &GetImageScanComponentReq{}
}

func (req *GetImageScanComponentReq) Validate() hperrors.ValidationErrors {
	if _, found := imagescanservice.ComponentByKey(req.Component); !found {
		return hperrors.NewValidationErrors(vld.Validate(vld.Must(false).OnError(
			vld.SetField("component", nil),
			vld.SetCustomKey("ERR_VLD_VALUE_INVALID"),
		)))
	}
	return hperrors.NewValidationErrors(vld.Validate())
}

type GetImageScanComponentResp struct {
	Meta *basedto.Meta           `json:"meta"`
	Data *ComponentImageScanResp `json:"data"`
}

// ComponentImageScanResp is what a part of HivePaaS ran when last scanned, as an
// app's Security tab shows it: after the global ignores, the only ones that
// apply to it.
type ComponentImageScanResp struct {
	Component *ComponentRefResp                 `json:"component"`
	Scan      *appimagescandto.AppImageScanResp `json:"scan"`
}
