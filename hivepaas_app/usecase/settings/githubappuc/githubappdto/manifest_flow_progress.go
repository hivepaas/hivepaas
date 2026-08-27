package githubappdto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
)

type HandleGithubAppManifestFlowProgressReq struct {
	SettingID      string `json:"-" mapstructure:"-"`
	Code           string `json:"-" mapstructure:"code"`
	State          string `json:"-" mapstructure:"state"`
	InstallationID int64  `json:"-" mapstructure:"installation_id"`
	SetupAction    string `json:"-" mapstructure:"setup_action"`
}

func NewHandleGithubAppManifestFlowProgressReq() *HandleGithubAppManifestFlowProgressReq {
	return &HandleGithubAppManifestFlowProgressReq{}
}

// Validate implements interface basedto.ReqValidator
func (req *HandleGithubAppManifestFlowProgressReq) Validate() hperrors.ValidationErrors {
	validators := make([]vld.Validator, 0, 10) //nolint:mnd
	return hperrors.NewValidationErrors(vld.Validate(validators...))
}

type HandleGithubAppManifestFlowProgressResp struct {
	Meta *basedto.Meta                                `json:"meta"`
	Data *HandleGithubAppManifestFlowProgressDataResp `json:"data"`
}

type HandleGithubAppManifestFlowProgressDataResp struct {
	RedirectURL string `json:"redirectURL"`
}
