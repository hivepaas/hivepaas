package imservicedto

import (
	vld "github.com/tiendc/go-validator"
	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
)

type TestSendInstantMsgReq struct {
	*IMServiceBaseReq
	TestMsg string `json:"testMsg"`
}

func NewTestSendInstantMsgReq() *TestSendInstantMsgReq {
	return &TestSendInstantMsgReq{}
}

func (req *TestSendInstantMsgReq) ModifyRequest() error {
	// NOTE: make sure req.Name is not empty to not fail the validation
	req.Name = gofn.Coalesce(req.Name, "x")
	req.TestMsg = gofn.Coalesce(req.TestMsg, "test message")
	return nil
}

// Validate implements interface basedto.ReqValidator
func (req *TestSendInstantMsgReq) Validate() hperrors.ValidationErrors {
	validators := make([]vld.Validator, 0, 10) //nolint:mnd
	validators = append(validators, req.validate("")...)
	return hperrors.NewValidationErrors(vld.Validate(validators...))
}

type TestSendInstantMsgResp struct {
	Meta *basedto.Meta `json:"meta"`
}
