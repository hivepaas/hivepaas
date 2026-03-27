package sessiondto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
)

const (
	minKeyLen = 10
	maxKeyLen = 100
)

type LoginWithAPIKeyReq struct {
	KeyID     string `json:"keyID"`
	SecretKey string `json:"secretKey"`
}

func NewLoginWithAPIKeyReq() *LoginWithAPIKeyReq {
	return &LoginWithAPIKeyReq{}
}

func (req *LoginWithAPIKeyReq) Validate() apperrors.ValidationErrors {
	var validators []vld.Validator
	validators = append(validators, basedto.ValidateStr(&req.KeyID, true, minKeyLen, maxKeyLen, "keyID")...)
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
