package githubappdto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
)

type TestGithubAppConnReq struct {
	*GithubAppBaseReq
}

func NewTestGithubAppConnReq() *TestGithubAppConnReq {
	return &TestGithubAppConnReq{}
}

func (req *TestGithubAppConnReq) ModifyRequest() error {
	return req.modifyRequest()
}

// Validate implements interface basedto.ReqValidator
func (req *TestGithubAppConnReq) Validate() apperrors.ValidationErrors {
	validators := make([]vld.Validator, 0, 5) //nolint:mnd
	validators = append(validators, req.validate("")...)
	return apperrors.NewValidationErrors(vld.Validate(validators...))
}

type TestGithubAppConnResp struct {
	Meta *basedto.Meta `json:"meta"`
}
