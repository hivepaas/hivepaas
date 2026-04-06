package volumedto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
)

type DeleteVolumeReq struct {
	VolumeID  string `json:"-"`
	ProjectID string `json:"-"`
	Force     bool   `json:"-" mapstructure:"force"`
}

func NewDeleteVolumeReq() *DeleteVolumeReq {
	return &DeleteVolumeReq{}
}

func (req *DeleteVolumeReq) Validate() apperrors.ValidationErrors {
	var validators []vld.Validator
	// NOTE: volume id is docker id, it's not ULID
	validators = append(validators, basedto.ValidateStr(&req.VolumeID, true, 1, volumeIDMaxLen, "volumeID")...)
	validators = append(validators, basedto.ValidateID(&req.ProjectID, false, "projectID")...)
	return apperrors.NewValidationErrors(vld.Validate(validators...))
}

type DeleteVolumeResp struct {
	Meta *basedto.Meta `json:"meta"`
}
