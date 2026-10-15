package appimagescandto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
)

type GetAppImageSBOMReq struct {
	ProjectID    string `json:"-"`
	ProjectEnvID string `json:"-"`
	AppID        string `json:"-"`
	// ImageID is an image the app's tasks run, as its Security tab lists it.
	ImageID string `json:"-" mapstructure:"imageId"`
}

func NewGetAppImageSBOMReq() *GetAppImageSBOMReq {
	return &GetAppImageSBOMReq{}
}

func (req *GetAppImageSBOMReq) Validate() hperrors.ValidationErrors {
	return hperrors.NewValidationErrors(vld.Validate(basedto.ValidateID(&req.ImageID, true, "imageId")...))
}

// GetAppImageSBOMResp is the image's SBOM, a file: its name and its content.
type GetAppImageSBOMResp struct {
	FileName string
	Content  []byte
}
