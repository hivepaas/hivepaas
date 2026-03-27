package volumedto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
)

type GetVolumeInspectionReq struct {
	VolumeID string `json:"-"`
}

func NewGetVolumeInspectionReq() *GetVolumeInspectionReq {
	return &GetVolumeInspectionReq{}
}

func (req *GetVolumeInspectionReq) Validate() apperrors.ValidationErrors {
	var validators []vld.Validator
	// NOTE: volume id is docker id, it's not ULID
	validators = append(validators, basedto.ValidateStr(&req.VolumeID, true, 1, volumeIDMaxLen, "volumeID")...)
	return apperrors.NewValidationErrors(vld.Validate(validators...))
}

type GetVolumeInspectionResp struct {
	Meta *basedto.Meta `json:"meta"`
	Data string        `json:"data"`
}
