package githubappdto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
)

type BeginGithubAppManifestFlowCreationReq struct {
	SettingID string `json:"-" mapstructure:"-"`
	State     string `json:"-" mapstructure:"state"`
}

func NewBeginGithubAppManifestFlowCreationReq() *BeginGithubAppManifestFlowCreationReq {
	return &BeginGithubAppManifestFlowCreationReq{}
}

// Validate implements interface basedto.ReqValidator
func (req *BeginGithubAppManifestFlowCreationReq) Validate() apperrors.ValidationErrors {
	var validators []vld.Validator
	return apperrors.NewValidationErrors(vld.Validate(validators...))
}

type BeginGithubAppManifestFlowCreationResp struct {
	Meta *basedto.Meta                               `json:"meta"`
	Data *BeginGithubAppManifestFlowCreationDataResp `json:"data"`
}

type BeginGithubAppManifestFlowCreationDataResp struct {
	PageContent string
}
