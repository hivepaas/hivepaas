package sslproviderdto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
)

const (
	eabKidMaxLen     = 100
	eabHmacKeyMaxLen = 200
)

type CreateSSLProviderReq struct {
	settings.CreateSettingReq
	*SSLProviderBaseReq
}

type SSLProviderBaseReq struct {
	Name           string           `json:"name"`
	Kind           base.SSLProvider `json:"kind"`
	Email          string           `json:"email"`
	DefaultKeyType base.SSLKeyType  `json:"defaultKeyType,omitempty"`

	LetsEncrypt *SSLProviderLetsEncryptReq `json:"letsEncrypt"`
	ZeroSSL     *SSLProviderZeroSSLReq     `json:"zeroSSL"`
	GoogleTrust *SSLProviderGoogleTrustReq `json:"googleTrust"`
}

func (req *SSLProviderBaseReq) ToEntity() *entity.SSLProvider {
	sslProvider := &entity.SSLProvider{
		Email:          req.Email,
		DefaultKeyType: req.DefaultKeyType,
	}
	switch req.Kind {
	case base.SSLProviderLetsEncrypt:
		sslProvider.LetsEncrypt = req.LetsEncrypt.ToEntity()
	case base.SSLProviderZeroSSL:
		sslProvider.ZeroSSL = req.ZeroSSL.ToEntity()
	case base.SSLProviderGoogleTrust:
		sslProvider.GoogleTrust = req.GoogleTrust.ToEntity()
	}
	return sslProvider
}

type SSLProviderLetsEncryptReq struct {
}

func (req *SSLProviderLetsEncryptReq) ToEntity() *entity.SSLProviderLetsEncrypt {
	return &entity.SSLProviderLetsEncrypt{}
}

func (req *SSLProviderLetsEncryptReq) validate(_ string) []vld.Validator {
	return nil
}

type SSLProviderZeroSSLReq struct {
	EABKid     string `json:"eabKid"`
	EABHmacKey string `json:"eabHmacKey"`
}

func (req *SSLProviderZeroSSLReq) ToEntity() *entity.SSLProviderZeroSSL {
	return &entity.SSLProviderZeroSSL{
		EABKid:     req.EABKid,
		EABHmacKey: entity.NewEncryptedField(req.EABHmacKey),
	}
}

func (req *SSLProviderZeroSSLReq) validate(field string) (res []vld.Validator) {
	if req == nil {
		return nil
	}
	if field != "" {
		field += "."
	}
	res = append(res, basedto.ValidateStr(&req.EABKid, true, 1, eabKidMaxLen, field+"eabKid")...)
	res = append(res, basedto.ValidateStr(&req.EABHmacKey, true, 1, eabHmacKeyMaxLen, field+"eabHmacKey")...)
	res = append(res, basedto.ValidatePlainSecret(&req.EABHmacKey, field+"eabHmacKey")...)
	return res
}

type SSLProviderGoogleTrustReq struct {
	EABKid     string `json:"eabKid"`
	EABHmacKey string `json:"eabHmacKey"`
}

func (req *SSLProviderGoogleTrustReq) ToEntity() *entity.SSLProviderGoogleTrust {
	return &entity.SSLProviderGoogleTrust{
		EABKid:     req.EABKid,
		EABHmacKey: entity.NewEncryptedField(req.EABHmacKey),
	}
}

func (req *SSLProviderGoogleTrustReq) validate(field string) (res []vld.Validator) {
	if req == nil {
		return nil
	}
	if field != "" {
		field += "."
	}
	res = append(res, basedto.ValidateStr(&req.EABKid, true, 1, eabKidMaxLen, field+"eabKid")...)
	res = append(res, basedto.ValidateStr(&req.EABHmacKey, true, 1, eabHmacKeyMaxLen, field+"eabHmacKey")...)
	res = append(res, basedto.ValidatePlainSecret(&req.EABHmacKey, field+"eabHmacKey")...)
	return res
}

// SecretFields lists the request's secret values in one place, so the paths that
// care about them do not each restate the list. Only the selected kind is listed:
// ToEntity ignores the others, so a placeholder left in them is never stored.
func (req *SSLProviderBaseReq) SecretFields() []basedto.SecretField {
	switch {
	case req.Kind == base.SSLProviderZeroSSL && req.ZeroSSL != nil:
		return []basedto.SecretField{{Path: "zeroSSL.eabHmacKey", Value: &req.ZeroSSL.EABHmacKey}}
	case req.Kind == base.SSLProviderGoogleTrust && req.GoogleTrust != nil:
		return []basedto.SecretField{{Path: "googleTrust.eabHmacKey", Value: &req.GoogleTrust.EABHmacKey}}
	}
	return nil
}

// KeepMaskedSecrets restores the stored values for the secrets the request only
// carries as the masked placeholder the GET response substitutes for them.
func (req *SSLProviderBaseReq) KeepMaskedSecrets(provider, current *entity.SSLProvider) {
	if current == nil {
		return
	}
	switch req.Kind {
	case base.SSLProviderZeroSSL:
		if req.ZeroSSL != nil && basedto.IsMaskedSecret(req.ZeroSSL.EABHmacKey) &&
			provider.ZeroSSL != nil && current.ZeroSSL != nil {
			provider.ZeroSSL.EABHmacKey = current.ZeroSSL.EABHmacKey
		}
	case base.SSLProviderGoogleTrust:
		if req.GoogleTrust != nil && basedto.IsMaskedSecret(req.GoogleTrust.EABHmacKey) &&
			provider.GoogleTrust != nil && current.GoogleTrust != nil {
			provider.GoogleTrust.EABHmacKey = current.GoogleTrust.EABHmacKey
		}
	case base.SSLProviderLetsEncrypt:
	}
}

func (req *SSLProviderBaseReq) validate(field string) (res []vld.Validator) {
	if field != "" {
		field += "."
	}
	switch req.Kind {
	case base.SSLProviderLetsEncrypt:
		res = append(res, basedto.ValidateCond(req.LetsEncrypt != nil, field+"letsEncrypt")...)
		res = append(res, req.LetsEncrypt.validate(field+"letsEncrypt")...)
	case base.SSLProviderZeroSSL:
		res = append(res, basedto.ValidateCond(req.ZeroSSL != nil, field+"zeroSSL")...)
		res = append(res, req.ZeroSSL.validate(field+"zeroSSL")...)
	case base.SSLProviderGoogleTrust:
		res = append(res, basedto.ValidateCond(req.GoogleTrust != nil, field+"googleTrust")...)
		res = append(res, req.GoogleTrust.validate(field+"googleTrust")...)
	}
	return res
}

func NewCreateSSLProviderReq() *CreateSSLProviderReq {
	return &CreateSSLProviderReq{}
}

// Validate implements interface basedto.ReqValidator
func (req *CreateSSLProviderReq) Validate() hperrors.ValidationErrors {
	validators := make([]vld.Validator, 0, 10) //nolint:mnd
	validators = append(validators, req.CreateSettingReq.Validate()...)
	validators = append(validators, req.validate("")...)
	// Creation has no stored value to fall back on, so the placeholder is not a
	// meaningful input here the way it is on update.
	validators = append(validators, basedto.ValidateNoMaskedSecrets(req.SecretFields())...)
	return hperrors.NewValidationErrors(vld.Validate(validators...))
}

type CreateSSLProviderResp struct {
	Meta *basedto.Meta         `json:"meta"`
	Data *basedto.ObjectIDResp `json:"data"`
}
