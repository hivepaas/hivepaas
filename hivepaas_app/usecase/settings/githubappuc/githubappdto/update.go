package githubappdto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
)

type UpdateGithubAppReq struct {
	settings.UpdateSettingReq
	*GithubAppBaseReq
}

func NewUpdateGithubAppReq() *UpdateGithubAppReq {
	return &UpdateGithubAppReq{}
}

// Validate implements interface basedto.ReqValidator
func (req *UpdateGithubAppReq) Validate() apperrors.ValidationErrors {
	var validators []vld.Validator
	validators = append(validators, req.validate("")...)
	return apperrors.NewValidationErrors(vld.Validate(validators...))
}

type UpdateGithubAppResp struct {
	Meta *basedto.Meta `json:"meta"`
}
