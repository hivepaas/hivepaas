package sessiondto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
)

const (
	minKeyLen = 10
	maxKeyLen = 100
)

type LoginWithAPIKeyReq struct {
	KeyID     string `json:"keyId"`
	SecretKey string `json:"secretKey"`
}

func NewLoginWithAPIKeyReq() *LoginWithAPIKeyReq {
	return &LoginWithAPIKeyReq{}
}

func (req *LoginWithAPIKeyReq) Validate() apperrors.ValidationErrors {
	var validators []vld.Validator
	validators = append(validators, basedto.ValidateStr(&req.KeyID, true, minKeyLen, maxKeyLen, "keyId")...)
	validators = append(validators, basedto.ValidateStr(&req.SecretKey, true, minKeyLen, maxKeyLen, "secretKey")...)
	return apperrors.NewValidationErrors(vld.Validate(validators...))
}

type LoginWithAPIKeyResp struct {
	Meta *basedto.Meta            `json:"meta"`
	Data *LoginWithAPIKeyDataResp `json:"data"`
}

type LoginWithAPIKeyDataResp struct {
	Session *BaseCreateSessionResp `json:"session"`
}
