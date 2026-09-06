package sshkeydto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
)

const (
	kindMaxLen       = 50
	privKeyMaxLen    = 10000
	pubKeyMaxLen     = 2000
	passphraseMaxLen = 100
)

type CreateSSHKeyReq struct {
	settings.CreateSettingReq
	*SSHKeyBaseReq
}

type SSHKeyBaseReq struct {
	Kind       base.SSHKeyKind     `json:"kind"`
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

// SecretFields lists the request's secret values in one place, so the paths that
// care about them do not each restate the list.
func (req *SSHKeyBaseReq) SecretFields() []basedto.SecretField {
	return []basedto.SecretField{
		{Path: "privateKey", Value: &req.PrivateKey},
		{Path: "passphrase", Value: &req.Passphrase},
	}
}

// KeepMaskedSecrets restores the stored values for the secrets the request only
// carries as the masked placeholder the GET response substitutes for them.
func (req *SSHKeyBaseReq) KeepMaskedSecrets(sshKey, current *entity.SSHKey) {
	if current == nil {
		return
	}
	if basedto.IsMaskedSecret(req.PrivateKey) {
		sshKey.PrivateKey = current.PrivateKey
	}
	if basedto.IsMaskedSecret(req.Passphrase) {
		sshKey.Passphrase = current.Passphrase
	}
}

func (req *SSHKeyBaseReq) validate(field string) (res []vld.Validator) {
	if field != "" {
		field += "."
	}
	res = append(res, basedto.ValidateStr(&req.Kind, true, 1, kindMaxLen, field+"kind")...)
	res = append(res, basedto.ValidateStr(&req.Name, true, 1, base.SettingNameMaxLen, field+"name")...)
	res = append(res, basedto.ValidateStr(&req.PublicKey, false, 1, pubKeyMaxLen, field+"publicKey")...)
	res = append(res, basedto.ValidateStr(&req.PrivateKey, true, 1, privKeyMaxLen, field+"privateKey")...)
	res = append(res, basedto.ValidatePlainSecret(&req.PrivateKey, field+"privateKey")...)
	res = append(res, basedto.ValidateStr(&req.Passphrase, false, 1, passphraseMaxLen, field+"passphrase")...)
	res = append(res, basedto.ValidatePlainSecret(&req.Passphrase, field+"passphrase")...)
	return res
}

func NewCreateSSHKeyReq() *CreateSSHKeyReq {
	return &CreateSSHKeyReq{}
}

// Validate implements interface basedto.ReqValidator
func (req *CreateSSHKeyReq) Validate() hperrors.ValidationErrors {
	validators := make([]vld.Validator, 0, 10) //nolint:mnd
	validators = append(validators, req.CreateSettingReq.Validate()...)
	validators = append(validators, req.validate("")...)
	// Creation has no stored value to fall back on, so the placeholder is not a
	// meaningful input here the way it is on update.
	validators = append(validators, basedto.ValidateNoMaskedSecrets(req.SecretFields())...)
	return hperrors.NewValidationErrors(vld.Validate(validators...))
}

type CreateSSHKeyResp struct {
	Meta *basedto.Meta         `json:"meta"`
	Data *basedto.ObjectIDResp `json:"data"`
}
