package sshkeydto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
)

const (
	privKeyMaxLen    = 10000
	pubKeyMaxLen     = 2000
	passphraseMaxLen = 100
)

type CreateSSHKeyReq struct {
	settings.CreateSettingReq
	*SSHKeyBaseReq
}

type SSHKeyBaseReq struct {
	Name       string              `json:"name"`
	KeyType    base.PrivateKeyType `json:"keyType"`
	PublicKey  string              `json:"publicKey"`
	PrivateKey string              `json:"privateKey"`
	Passphrase string              `json:"passphrase"`
}

func (req *SSHKeyBaseReq) ToEntity() *entity.SSHKey {
	return &entity.SSHKey{
		KeyType:    req.KeyType,
		PublicKey:  req.PublicKey,
		PrivateKey: entity.NewEncryptedField(req.PrivateKey),
		Passphrase: entity.NewEncryptedField(req.Passphrase),
	}
}

func (req *SSHKeyBaseReq) validate(field string) (res []vld.Validator) {
	if field != "" {
		field += "."
	}
	res = append(res, basedto.ValidateStr(&req.Name, true, 1, base.SettingNameMaxLen, field+"name")...)
	res = append(res, basedto.ValidateStr(&req.PublicKey, false, 1, pubKeyMaxLen, field+"publicKey")...)
	res = append(res, basedto.ValidateStr(&req.PrivateKey, true, 1, privKeyMaxLen, field+"privateKey")...)
	res = append(res, basedto.ValidateStr(&req.Passphrase, false, 1, passphraseMaxLen, field+"passphrase")...)
	return res
}

func NewCreateSSHKeyReq() *CreateSSHKeyReq {
	return &CreateSSHKeyReq{}
}

// Validate implements interface basedto.ReqValidator
func (req *CreateSSHKeyReq) Validate() apperrors.ValidationErrors {
	var validators []vld.Validator
	validators = append(validators, req.validate("")...)
	return apperrors.NewValidationErrors(vld.Validate(validators...))
}

type CreateSSHKeyResp struct {
	Meta *basedto.Meta         `json:"meta"`
	Data *basedto.ObjectIDResp `json:"data"`
}
