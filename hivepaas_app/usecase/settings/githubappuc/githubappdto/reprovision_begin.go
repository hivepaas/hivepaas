package githubappdto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
)

type BeginReprovisionGithubAppReq struct {
	settings.BaseSettingReq
	ID        string `json:"-"`
	Name      string `json:"name"`
	UpdateVer int    `json:"updateVer"`
}

func NewBeginReprovisionGithubAppReq() *BeginReprovisionGithubAppReq {
	return &BeginReprovisionGithubAppReq{}
}

// Validate implements interface basedto.ReqValidator
func (req *BeginReprovisionGithubAppReq) Validate() apperrors.ValidationErrors {
	var validators []vld.Validator
	return apperrors.NewValidationErrors(vld.Validate(validators...))
}

type BeginReprovisionGithubAppResp struct {
	Meta *basedto.Meta                      `json:"meta"`
	Data *BeginReprovisionGithubAppDataResp `json:"data"`
}

type BeginReprovisionGithubAppDataResp struct {
	RedirectURL string `json:"redirectURL"`
}
