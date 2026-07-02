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
	var validators []vld.Validator
	validators = append(validators, req.validate("")...)
	return apperrors.NewValidationErrors(vld.Validate(validators...))
}

type TestGithubAppConnResp struct {
	Meta *basedto.Meta `json:"meta"`
}
