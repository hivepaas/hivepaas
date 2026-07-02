package githubappdto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
)

type UpdateGithubAppStatusReq struct {
	settings.UpdateSettingStatusReq
}

func NewUpdateGithubAppStatusReq() *UpdateGithubAppStatusReq {
	return &UpdateGithubAppStatusReq{}
}

// Validate implements interface basedto.ReqValidator
func (req *UpdateGithubAppStatusReq) Validate() apperrors.ValidationErrors {
	var validators []vld.Validator
	validators = append(validators, basedto.ValidateStrIn(req.Status, false,
		base.AllSettingSettableStatuses, "status")...)
	return apperrors.NewValidationErrors(vld.Validate(validators...))
}

type UpdateGithubAppStatusResp struct {
	Meta *basedto.Meta `json:"meta"`
}
