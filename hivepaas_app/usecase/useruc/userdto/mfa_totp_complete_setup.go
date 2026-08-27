package userdto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
)

const (
	minPasscodeLen = 1
	maxPasscodeLen = 10

	minTokenLen = 10
	maxTokenLen = 10000
)

type CompleteMFATotpSetupReq struct {
	Passcode  string `json:"passcode"`
	TotpToken string `json:"totpToken"`
}

func NewCompleteMFATotpSetupReq() *CompleteMFATotpSetupReq {
	return &CompleteMFATotpSetupReq{}
}

func (req *CompleteMFATotpSetupReq) Validate() hperrors.ValidationErrors {
	validators := make([]vld.Validator, 0, 10) //nolint:mnd
	validators = append(validators, basedto.ValidateStr(&req.Passcode, true,
		minPasscodeLen, maxPasscodeLen, "passcode")...)
	validators = append(validators, basedto.ValidateStr(&req.TotpToken, true,
		minTokenLen, maxTokenLen, "totpToken")...)
	return hperrors.NewValidationErrors(vld.Validate(validators...))
}

type CompleteMFATotpSetupResp struct {
	Meta *basedto.Meta `json:"meta"`
}
