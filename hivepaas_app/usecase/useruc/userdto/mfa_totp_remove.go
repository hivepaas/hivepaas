package userdto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
)

type RemoveMFATotpReq struct {
	Passcode string `json:"passcode"`
}

func NewRemoveMFATotpReq() *RemoveMFATotpReq {
	return &RemoveMFATotpReq{}
}

func (req *RemoveMFATotpReq) Validate() apperrors.ValidationErrors {
	var validators []vld.Validator
	validators = append(validators, basedto.ValidateStr(&req.Passcode, true,
		minPasscodeLen, maxPasscodeLen, "passcode")...)
	return apperrors.NewValidationErrors(vld.Validate(validators...))
}

type RemoveMFATotpResp struct {
	Meta *basedto.Meta `json:"meta"`
}
