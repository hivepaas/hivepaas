package sslcertdto

import (
	"time"

	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/copier"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
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
	validators := make([]vld.Validator, 0, 10) //nolint:mnd
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
	Provider      *settings.BaseSettingResp          `json:"provider"`
	Domain        string                             `json:"domain"`
	Certificate   string                             `json:"certificate"`
	PrivateKey    string                             `json:"privateKey"`
	CACertificate string                             `json:"caCertificate,omitempty"`
	KeyType       base.SSLKeyType                    `json:"keyType"`
	ValidPeriod   timeutil.Duration                  `json:"validPeriod"`
	Email         string                             `json:"email"`
	AutoRenew     bool                               `json:"autoRenew"`
	AcmeProvider  *settings.BaseSettingResp          `json:"acmeProvider"`
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

	if config.Provider.ID != "" {
		providerSetting := refObjects.RefSettings[config.Provider.ID]
		providerResp, _ := settings.TransformSettingBase(providerSetting)
		if providerResp == nil {
			providerResp = settings.NewMissingSetting(config.Provider.ID, base.SettingTypeSSLProvider)
		}
		resp.Provider = providerResp
	} else {
		resp.Provider = nil
	}

	if config.AcmeProvider.ID != "" {
		providerSetting := refObjects.RefSettings[config.AcmeProvider.ID]
		acmeProviderResp, _ := settings.TransformSettingBase(providerSetting)
		if acmeProviderResp == nil {
			acmeProviderResp = settings.NewMissingSetting(config.AcmeProvider.ID, base.SettingTypeAcmeDnsProvider)
		}
		resp.AcmeProvider = acmeProviderResp
	} else {
		resp.AcmeProvider = nil
	}

	resp.SecretMasked = config.PrivateKey.IsEncrypted() || resp.Inherited
	if resp.SecretMasked {
		resp.PrivateKey = maskedSecret
		resp.Certificate = maskedSecret
		if resp.CACertificate != "" {
			resp.CACertificate = maskedSecret
		}
	}

	resp.Notification = basedto.TransformBaseEventNotification(config.Notification, refObjects)
	return resp, nil
}

func TransformSSLCertBasic(
	setting *entity.Setting,
	refObjects *entity.RefObjects,
) (*SSLCertResp, error) {
	if setting == nil {
		return nil, nil
	}
	resp, err := TransformSSLCert(setting, refObjects)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}
	resp.Certificate = maskedSecret
	resp.PrivateKey = maskedSecret
	if resp.CACertificate != "" {
		resp.CACertificate = maskedSecret
	}
	resp.SecretMasked = true
	resp.Notification = nil
	return resp, nil
}
