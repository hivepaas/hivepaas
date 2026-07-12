package oauthdto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/copier"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
)

const (
	maskedSecret = "****************"
)

type GetOAuthReq struct {
	settings.GetSettingReq
}

func NewGetOAuthReq() *GetOAuthReq {
	return &GetOAuthReq{}
}

func (req *GetOAuthReq) Validate() apperrors.ValidationErrors {
	var validators []vld.Validator
	validators = append(validators, req.GetSettingReq.Validate()...)
	return apperrors.NewValidationErrors(vld.Validate(validators...))
}

type GetOAuthResp struct {
	Meta *basedto.Meta `json:"meta"`
	Data *OAuthResp    `json:"data"`
}

type OAuthResp struct {
	*settings.BaseSettingResp
	ClientID         string   `json:"clientId"`
	ClientSecret     string   `json:"clientSecret"`
	Organization     string   `json:"organization"`
	CallbackURL      string   `json:"callbackURL"`
	AuthURL          string   `json:"authURL,omitempty"`
	TokenURL         string   `json:"tokenURL,omitempty"`
	ProfileURL       string   `json:"profileURL,omitempty"`
	AutoDiscoveryURL string   `json:"autoDiscoveryURL,omitempty"`
	Scopes           []string `json:"scopes,omitempty"`
	SecretMasked     bool     `json:"secretMasked,omitempty"`
}

func (resp *OAuthResp) CopyClientSecret(field entity.EncryptedField) error {
	resp.ClientSecret = field.String()
	return nil
}

type OAuthTransformInput struct {
	RefObjects      *entity.RefObjects
	BaseCallbackURL string
}

func TransformOAuth(
	setting *entity.Setting,
	input *OAuthTransformInput,
) (resp *OAuthResp, err error) {
	config := setting.MustAsOAuth()
	if err = copier.Copy(&resp, config); err != nil {
		return nil, apperrors.Wrap(err)
	}

	resp.BaseSettingResp, err = settings.TransformSettingBase(setting)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	// Recalculate callbackURL for the oauth as it depends on the actual server address
	resp.CallbackURL = input.BaseCallbackURL + "/" + setting.ID
	resp.SecretMasked = config.ClientSecret.IsEncrypted() || resp.Inherited
	if resp.SecretMasked {
		resp.ClientSecret = maskedSecret
	}

	return resp, nil
}
