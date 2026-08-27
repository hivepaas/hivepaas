package sessiondto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
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

func (req *LoginWithAPIKeyReq) Validate() hperrors.ValidationErrors {
	validators := make([]vld.Validator, 0, 10) //nolint:mnd
	validators = append(validators, basedto.ValidateStr(&req.KeyID, true, minKeyLen, maxKeyLen, "keyId")...)
	validators = append(validators, basedto.ValidateStr(&req.SecretKey, true, minKeyLen, maxKeyLen, "secretKey")...)
	return hperrors.NewValidationErrors(vld.Validate(validators...))
}

type LoginWithAPIKeyResp struct {
	Meta *basedto.Meta            `json:"meta"`
	Data *LoginWithAPIKeyDataResp `json:"data"`
}

type LoginWithAPIKeyDataResp struct {
	Session *BaseCreateSessionResp `json:"session"`
}
