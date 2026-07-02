package sessiondto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
)

const (
	maxPasscodeLen = 10
	maxMFATokenLen = 10000
)

type LoginWithPasscodeReq struct {
	Passcode string `json:"passcode"`
	MFAToken string `json:"mfaToken"`
}

func NewLoginWithPasscodeReq() *LoginWithPasscodeReq {
	return &LoginWithPasscodeReq{}
}

func (req *LoginWithPasscodeReq) Validate() apperrors.ValidationErrors {
	var validators []vld.Validator
	validators = append(validators, basedto.ValidateStr(&req.Passcode, true, 1,
		maxPasscodeLen, "passcode")...)
	validators = append(validators, basedto.ValidateStr(&req.MFAToken, true, 1,
		maxMFATokenLen, "mfaToken")...)
	return apperrors.NewValidationErrors(vld.Validate(validators...))
}

type LoginWithPasscodeResp struct {
	Meta *basedto.Meta              `json:"meta"`
	Data *LoginWithPasscodeDataResp `json:"data"`
}

type LoginWithPasscodeDataResp struct {
	Session *BaseCreateSessionResp `json:"session,omitempty"`
}
