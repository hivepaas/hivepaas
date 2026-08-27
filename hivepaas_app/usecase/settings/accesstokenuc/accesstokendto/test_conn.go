package accesstokendto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
)

type TestAccessTokenConnReq struct {
	*AccessTokenBaseReq
}

func NewTestAccessTokenConnReq() *TestAccessTokenConnReq {
	return &TestAccessTokenConnReq{}
}

func (req *TestAccessTokenConnReq) ModifyRequest() error {
	return req.modifyRequest()
}

// Validate implements interface basedto.ReqValidator
func (req *TestAccessTokenConnReq) Validate() hperrors.ValidationErrors {
	validators := make([]vld.Validator, 0, 10) //nolint:mnd
	validators = append(validators, req.validate("")...)
	return hperrors.NewValidationErrors(vld.Validate(validators...))
}

type TestAccessTokenConnResp struct {
	Meta *basedto.Meta `json:"meta"`
}
