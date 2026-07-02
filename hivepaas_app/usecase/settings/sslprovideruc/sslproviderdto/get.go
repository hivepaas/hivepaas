package sslproviderdto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/copier"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
)

const (
	maskedSecret = "****************"
)

type GetSSLProviderReq struct {
	settings.GetSettingReq
}

func NewGetSSLProviderReq() *GetSSLProviderReq {
	return &GetSSLProviderReq{}
}

func (req *GetSSLProviderReq) Validate() apperrors.ValidationErrors {
	var validators []vld.Validator
	validators = append(validators, req.GetSettingReq.Validate()...)
	return apperrors.NewValidationErrors(vld.Validate(validators...))
}

type GetSSLProviderResp struct {
	Meta *basedto.Meta    `json:"meta"`
	Data *SSLProviderResp `json:"data"`
}

type SSLProviderResp struct {
	*settings.BaseSettingResp
	Kind           base.SSLProvider `json:"kind"`
	Email          string           `json:"email"`
	DefaultKeyType base.SSLKeyType  `json:"defaultKeyType"`

	LetsEncrypt  *SSLProviderLetsEncryptResp `json:"letsEncrypt,omitempty"`
	ZeroSSL      *SSLProviderZeroSSLResp     `json:"zeroSSL,omitempty"`
	GoogleTrust  *SSLProviderGoogleTrustResp `json:"googleTrust,omitempty"`
	SecretMasked bool                        `json:"secretMasked,omitempty"`
}

type SSLProviderLetsEncryptResp struct {
}

type SSLProviderZeroSSLResp struct {
	EABKid     string `json:"eabKid"`
	EABHmacKey string `json:"eabHmacKey"`
}

func (resp *SSLProviderZeroSSLResp) CopyEABHmacKey(field entity.EncryptedField) error {
	resp.EABHmacKey = field.String()
	return nil
}

type SSLProviderGoogleTrustResp struct {
	EABKid     string `json:"eabKid"`
	EABHmacKey string `json:"eabHmacKey"`
}

func (resp *SSLProviderGoogleTrustResp) CopyEABHmacKey(field entity.EncryptedField) error {
	resp.EABHmacKey = field.String()
	return nil
}

func TransformSSLProvider(
	setting *entity.Setting,
	_ *entity.RefObjects,
) (resp *SSLProviderResp, err error) {
	config := setting.MustAsSSLProvider()
	if err = copier.Copy(&resp, config); err != nil {
		return nil, apperrors.New(err)
	}
	resp.Kind = base.SSLProvider(setting.Kind)

	resp.BaseSettingResp, err = settings.TransformSettingBase(setting)
	if err != nil {
		return nil, apperrors.New(err)
	}

	switch {
	case config.LetsEncrypt != nil:
		resp.SecretMasked = false
	case config.ZeroSSL != nil:
		resp.SecretMasked = config.ZeroSSL.EABHmacKey.IsEncrypted() || resp.Inherited
		if resp.SecretMasked {
			resp.ZeroSSL.EABHmacKey = maskedSecret
		}
	case config.GoogleTrust != nil:
		resp.SecretMasked = config.GoogleTrust.EABHmacKey.IsEncrypted() || resp.Inherited
		if resp.SecretMasked {
			resp.GoogleTrust.EABHmacKey = maskedSecret
		}
	}

	return resp, nil
}
