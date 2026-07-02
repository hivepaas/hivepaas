package emaildto

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

type GetEmailReq struct {
	settings.GetSettingReq
}

func NewGetEmailReq() *GetEmailReq {
	return &GetEmailReq{}
}

func (req *GetEmailReq) Validate() apperrors.ValidationErrors {
	var validators []vld.Validator
	validators = append(validators, req.GetSettingReq.Validate()...)
	return apperrors.NewValidationErrors(vld.Validate(validators...))
}

type GetEmailResp struct {
	Meta *basedto.Meta `json:"meta"`
	Data *EmailResp    `json:"data"`
}

type EmailResp struct {
	*settings.BaseSettingResp
	Kind         base.EmailKind `json:"kind"`
	SMTP         *EmailSMTPResp `json:"smtp,omitempty"`
	HTTP         *EmailHTTPResp `json:"http,omitempty"`
	SecretMasked bool           `json:"secretMasked,omitempty"`
}

type EmailSMTPResp struct {
	Host        string `json:"host"`
	Port        int    `json:"port"`
	Username    string `json:"username"`
	DisplayName string `json:"displayName"`
	Password    string `json:"password"`
	SSL         bool   `json:"ssl"`
}

func (resp *EmailSMTPResp) CopyPassword(field entity.EncryptedField) error {
	resp.Password = field.String()
	return nil
}

type EmailHTTPResp struct {
	Endpoint     string                        `json:"endpoint"`
	Method       base.HTTPMethod               `json:"method"`
	ContentType  string                        `json:"contentType"`
	Headers      map[string]string             `json:"headers"`
	FieldMapping *entity.EmailHTTPFieldMapping `json:"fieldMapping"`
	Username     string                        `json:"username"`
	DisplayName  string                        `json:"displayName"`
	Password     string                        `json:"password"`
}

func (resp *EmailHTTPResp) CopyPassword(field entity.EncryptedField) error {
	resp.Password = field.String()
	return nil
}

type EmailGmailAPIKeyResp struct {
	User   string `json:"user"`
	APIKey string `json:"apiKey"`
}

func (resp *EmailGmailAPIKeyResp) CopyPassword(field entity.EncryptedField) error {
	resp.APIKey = field.String()
	return nil
}

func TransformEmail(
	setting *entity.Setting,
	_ *entity.RefObjects,
) (resp *EmailResp, err error) {
	config := setting.MustAsEmail()
	if err = copier.Copy(&resp, config); err != nil {
		return nil, apperrors.New(err)
	}
	resp.Kind = base.EmailKind(setting.Kind)

	resp.BaseSettingResp, err = settings.TransformSettingBase(setting)
	if err != nil {
		return nil, apperrors.New(err)
	}

	switch {
	case config.SMTP != nil:
		resp.SecretMasked = config.SMTP.Password.IsEncrypted() || resp.Inherited
		if resp.SecretMasked {
			resp.SMTP.Password = maskedSecret
		}
	case config.HTTP != nil:
		resp.SecretMasked = config.HTTP.Password.IsEncrypted() || resp.Inherited
		if resp.SecretMasked {
			resp.HTTP.Password = maskedSecret
		}
	}

	return resp, nil
}
