package imservicedto

import (
	vld "github.com/tiendc/go-validator"
	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
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
func (req *TestSendInstantMsgReq) Validate() apperrors.ValidationErrors {
	var validators []vld.Validator
	validators = append(validators, req.validate("")...)
	return apperrors.NewValidationErrors(vld.Validate(validators...))
}

type TestSendInstantMsgResp struct {
	Meta *basedto.Meta `json:"meta"`
}
