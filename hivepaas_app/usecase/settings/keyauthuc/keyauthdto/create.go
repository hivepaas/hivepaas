package keyauthdto

import (
	"strings"

	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
)

const (
	// A key id is short everywhere - AWS's is 20 characters - but some stores
	// use longer ones; a secret key can be a long token.
	keyIDMaxLen     = 255
	secretKeyMaxLen = 1024
)

type CreateKeyAuthReq struct {
	settings.CreateSettingReq
	*KeyAuthBaseReq
}

type KeyAuthBaseReq struct {
	Name      string `json:"name"`
	KeyID     string `json:"keyId"`
	SecretKey string `json:"secretKey"`
}

func (req *KeyAuthBaseReq) ToEntity() *entity.KeyAuth {
	return &entity.KeyAuth{
		KeyID:     req.KeyID,
		SecretKey: entity.NewEncryptedField(req.SecretKey),
	}
}

// SecretFields lists the request's secret values in one place, so the paths that
// care about them do not each restate the list.
func (req *KeyAuthBaseReq) SecretFields() []basedto.SecretField {
	return []basedto.SecretField{{Path: "secretKey", Value: &req.SecretKey}}
}

// KeepMaskedSecrets restores the stored values for the secrets the request only
// carries as the masked placeholder the GET response substitutes for them.
func (req *KeyAuthBaseReq) KeepMaskedSecrets(keyAuth, current *entity.KeyAuth) {
	if current == nil {
		return
	}
	if basedto.IsMaskedSecret(req.SecretKey) {
		keyAuth.SecretKey = current.SecretKey
	}
}

func (req *KeyAuthBaseReq) modifyRequest() error {
	req.Name = strings.TrimSpace(req.Name)
	req.KeyID = strings.TrimSpace(req.KeyID)
	return nil
}

func (req *KeyAuthBaseReq) validate(field string) (res []vld.Validator) {
	if field != "" {
		field += "."
	}
	res = append(res, basedto.ValidateStr(&req.Name, true, 1, base.SettingNameMaxLen, field+"name")...)
	res = append(res, basedto.ValidateStr(&req.KeyID, true, 1, keyIDMaxLen, field+"keyId")...)
	res = append(res, basedto.ValidateStr(&req.SecretKey, true, 1, secretKeyMaxLen, field+"secretKey")...)
	res = append(res, basedto.ValidatePlainSecret(&req.SecretKey, field+"secretKey")...)
	return res
}

func NewCreateKeyAuthReq() *CreateKeyAuthReq {
	return &CreateKeyAuthReq{}
}

func (req *CreateKeyAuthReq) ModifyRequest() error {
	return req.modifyRequest()
}

// Validate implements interface basedto.ReqValidator
func (req *CreateKeyAuthReq) Validate() hperrors.ValidationErrors {
	validators := make([]vld.Validator, 0, 10) //nolint:mnd
	validators = append(validators, req.CreateSettingReq.Validate()...)
	validators = append(validators, req.validate("")...)
	// Creation has no stored value to fall back on, so the placeholder is not a
	// meaningful input here the way it is on update.
	validators = append(validators, basedto.ValidateNoMaskedSecrets(req.SecretFields())...)
	return hperrors.NewValidationErrors(vld.Validate(validators...))
}

type CreateKeyAuthResp struct {
	Meta *basedto.Meta         `json:"meta"`
	Data *basedto.ObjectIDResp `json:"data"`
}
