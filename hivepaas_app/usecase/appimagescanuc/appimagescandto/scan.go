package appimagescandto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
)

type ScanAppImagesReq struct {
	ProjectID    string `json:"-"`
	ProjectEnvID string `json:"-"`
	AppID        string `json:"-"`
}

func NewScanAppImagesReq() *ScanAppImagesReq {
	return &ScanAppImagesReq{}
}

func (req *ScanAppImagesReq) Validate() hperrors.ValidationErrors {
	validators := make([]vld.Validator, 0, 10) //nolint:mnd
	validators = append(validators, basedto.ValidateID(&req.ProjectID, true, "projectId")...)
	validators = append(validators, basedto.ValidateID(&req.ProjectEnvID, true, "projectEnv")...)
	validators = append(validators, basedto.ValidateID(&req.AppID, true, "appId")...)
	return hperrors.NewValidationErrors(vld.Validate(validators...))
}

type ScanAppImagesResp struct {
	Meta *basedto.Meta          `json:"meta"`
	Data *ScanAppImagesDataResp `json:"data"`
}

type ScanAppImagesDataResp struct {
	// Task is the scan: the one just queued, or one queued before that has not
	// started yet.
	Task *basedto.ObjectIDResp `json:"task"`
}
