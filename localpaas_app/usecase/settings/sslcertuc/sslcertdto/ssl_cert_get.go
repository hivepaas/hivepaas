package sslcertdto

import (
	"time"

	vld "github.com/tiendc/go-validator"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/base"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/entity"
	"github.com/localpaas/localpaas/localpaas_app/pkg/copier"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings"
)

const (
	maskedSecret = "****************"
)

type GetSSLCertReq struct {
	settings.GetSettingReq
}

func NewGetSSLCertReq() *GetSSLCertReq {
	return &GetSSLCertReq{}
}

func (req *GetSSLCertReq) Validate() apperrors.ValidationErrors {
	var validators []vld.Validator
	validators = append(validators, req.GetSettingReq.Validate()...)
	return apperrors.NewValidationErrors(vld.Validate(validators...))
}

type GetSSLCertResp struct {
	Meta *basedto.Meta `json:"meta"`
	Data *SSLCertResp  `json:"data"`
}

type SSLCertResp struct {
	*settings.BaseSettingResp
	CertType      base.SSLCertType                   `json:"certType"`
	Domain        string                             `json:"domain"`
	Certificate   string                             `json:"certificate"`
	PrivateKey    string                             `json:"privateKey"`
	KeyType       base.SSLKeyType                    `json:"keyType"`
	Email         string                             `json:"email"`
	AutoRenew     bool                               `json:"autoRenew"`
	RenewableFrom *time.Time                         `json:"renewableFrom" copy:",nilonzero"`
	ExpireAt      *time.Time                         `json:"expireAt" copy:",nilonzero"`
	NotifyFrom    *time.Time                         `json:"notifyFrom" copy:",nilonzero"`
	Notification  *basedto.BaseEventNotificationResp `json:"notification"`
	SecretMasked  bool                               `json:"secretMasked,omitempty"`
}

func (resp *SSLCertResp) CopyPrivateKey(field entity.EncryptedField) error {
	resp.PrivateKey = field.String()
	return nil
}

func TransformSSLCert(
	setting *entity.Setting,
	refObjects *entity.RefObjects,
) (resp *SSLCertResp, err error) {
	config := setting.MustAsSSLCert()
	if err = copier.Copy(&resp, config); err != nil {
		return nil, apperrors.Wrap(err)
	}

	resp.BaseSettingResp, err = settings.TransformSettingBase(setting)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	resp.SecretMasked = config.PrivateKey.IsEncrypted() || resp.Inherited
	if resp.SecretMasked {
		resp.PrivateKey = maskedSecret
	}

	resp.Notification = basedto.TransformBaseEventNotification(config.Notification, refObjects)
	return resp, nil
}

func TransformSSLCertNoCertData(
	setting *entity.Setting,
	refObjects *entity.RefObjects,
) (*SSLCertResp, error) {
	resp, err := TransformSSLCert(setting, refObjects)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}
	resp.Certificate = maskedSecret
	resp.PrivateKey = maskedSecret
	resp.SecretMasked = true
	return resp, nil
}
