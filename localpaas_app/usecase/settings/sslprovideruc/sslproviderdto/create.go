package sslproviderdto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/base"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/entity"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings"
)

const (
	eabKidMaxLen     = 100
	eabHmacKeyMaxLen = 100
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
	GoogleTS    *SSLProviderGoogleTSReq    `json:"googleTS"`
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
	case base.SSLProviderGoogleTS:
		sslProvider.GoogleTS = req.GoogleTS.ToEntity()
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
	return res
}

type SSLProviderGoogleTSReq struct {
	EABKid     string `json:"eabKid"`
	EABHmacKey string `json:"eabHmacKey"`
}

func (req *SSLProviderGoogleTSReq) ToEntity() *entity.SSLProviderGoogleTS {
	return &entity.SSLProviderGoogleTS{
		EABKid:     req.EABKid,
		EABHmacKey: entity.NewEncryptedField(req.EABHmacKey),
	}
}

func (req *SSLProviderGoogleTSReq) validate(field string) (res []vld.Validator) {
	if req == nil {
		return nil
	}
	if field != "" {
		field += "."
	}
	res = append(res, basedto.ValidateStr(&req.EABKid, true, 1, eabKidMaxLen, field+"eabKid")...)
	res = append(res, basedto.ValidateStr(&req.EABHmacKey, true, 1, eabHmacKeyMaxLen, field+"eabHmacKey")...)
	return res
}

func (req *SSLProviderBaseReq) validate(field string) (res []vld.Validator) {
	if field != "" {
		field += "."
	}
	switch req.Kind {
	case base.SSLProviderLetsEncrypt:
		res = append(res, basedto.ValidateValue(req.LetsEncrypt != nil, field+"letsEncrypt")...)
		res = append(res, req.LetsEncrypt.validate(field+"letsEncrypt")...)
	case base.SSLProviderZeroSSL:
		res = append(res, basedto.ValidateValue(req.ZeroSSL != nil, field+"zeroSSL")...)
		res = append(res, req.ZeroSSL.validate(field+"zeroSSL")...)
	case base.SSLProviderGoogleTS:
		res = append(res, basedto.ValidateValue(req.GoogleTS != nil, field+"googleTS")...)
		res = append(res, req.GoogleTS.validate(field+"googleTS")...)
	}
	return res
}

func NewCreateSSLProviderReq() *CreateSSLProviderReq {
	return &CreateSSLProviderReq{}
}

// Validate implements interface basedto.ReqValidator
func (req *CreateSSLProviderReq) Validate() apperrors.ValidationErrors {
	var validators []vld.Validator
	validators = append(validators, req.validate("")...)
	return apperrors.NewValidationErrors(vld.Validate(validators...))
}

type CreateSSLProviderResp struct {
	Meta *basedto.Meta         `json:"meta"`
	Data *basedto.ObjectIDResp `json:"data"`
}
