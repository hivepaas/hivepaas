package githubappdto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
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
func (req *UpdateGithubAppReq) Validate() hperrors.ValidationErrors {
	validators := make([]vld.Validator, 0, 10) //nolint:mnd
	validators = append(validators, req.validate("")...)
	return hperrors.NewValidationErrors(vld.Validate(validators...))
}

type UpdateGithubAppResp struct {
	Meta *basedto.Meta `json:"meta"`
}
