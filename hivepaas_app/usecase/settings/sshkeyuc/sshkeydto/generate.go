package sshkeydto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
)

type GenerateSSHKeyReq struct {
	KeyType    base.PrivateKeyType `json:"keyType"`
	Passphrase string              `json:"passphrase"`
}

func NewGenerateSSHKeyReq() *GenerateSSHKeyReq {
	return &GenerateSSHKeyReq{}
}

// Validate implements interface basedto.ReqValidator
func (req *GenerateSSHKeyReq) Validate() apperrors.ValidationErrors {
	var validators []vld.Validator
	validators = append(validators, basedto.ValidateStrIn(&req.KeyType, true,
		base.AllPrivateKeyTypes, "keyType")...)
	validators = append(validators, basedto.ValidateStr(&req.Passphrase, false, 1,
		passphraseMaxLen, "passphrase")...)
	return apperrors.NewValidationErrors(vld.Validate(validators...))
}

type GenerateSSHKeyResp struct {
	Meta *basedto.Meta           `json:"meta"`
	Data *GenerateSSHKeyDataResp `json:"data"`
}

type GenerateSSHKeyDataResp struct {
	KeyType    base.PrivateKeyType `json:"keyType"`
	PublicKey  string              `json:"publicKey"`
	PrivateKey string              `json:"privateKey"`
}
