package githubappdto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
)

type BeginGithubAppManifestFlowCreationReq struct {
	SettingID string `json:"-" mapstructure:"-"`
	State     string `json:"-" mapstructure:"state"`
}

func NewBeginGithubAppManifestFlowCreationReq() *BeginGithubAppManifestFlowCreationReq {
	return &BeginGithubAppManifestFlowCreationReq{}
}

// Validate implements interface basedto.ReqValidator
func (req *BeginGithubAppManifestFlowCreationReq) Validate() hperrors.ValidationErrors {
	validators := make([]vld.Validator, 0, 10) //nolint:mnd
	return hperrors.NewValidationErrors(vld.Validate(validators...))
}

type BeginGithubAppManifestFlowCreationResp struct {
	Meta *basedto.Meta                               `json:"meta"`
	Data *BeginGithubAppManifestFlowCreationDataResp `json:"data"`
}

type BeginGithubAppManifestFlowCreationDataResp struct {
	PageContent string
}
