package cloudstoragedto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
)

type UpdateCloudStorageReq struct {
	settings.UpdateSettingReq
	*CloudStorageBaseReq
}

func NewUpdateCloudStorageReq() *UpdateCloudStorageReq {
	return &UpdateCloudStorageReq{}
}

// Validate implements interface basedto.ReqValidator
func (req *UpdateCloudStorageReq) Validate() apperrors.ValidationErrors {
	var validators []vld.Validator
	validators = append(validators, req.validate("")...)
	return apperrors.NewValidationErrors(vld.Validate(validators...))
}

type UpdateCloudStorageResp struct {
	Meta *basedto.Meta `json:"meta"`
}
