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
	validators := make([]vld.Validator, 0, 5) //nolint:mnd
	return apperrors.NewValidationErrors(vld.Validate(validators...))
}

type BeginGithubAppManifestFlowCreationResp struct {
	Meta *basedto.Meta                               `json:"meta"`
	Data *BeginGithubAppManifestFlowCreationDataResp `json:"data"`
}

type BeginGithubAppManifestFlowCreationDataResp struct {
	PageContent string
}
