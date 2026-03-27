package githubappdto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/entity"
	"github.com/localpaas/localpaas/localpaas_app/pkg/copier"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings"
)

const (
	maskedSecret = "****************"
)

type GetGithubAppReq struct {
	settings.GetSettingReq
}

func NewGetGithubAppReq() *GetGithubAppReq {
	return &GetGithubAppReq{}
}

func (req *GetGithubAppReq) Validate() apperrors.ValidationErrors {
	var validators []vld.Validator
	validators = append(validators, req.GetSettingReq.Validate()...)
	return apperrors.NewValidationErrors(vld.Validate(validators...))
}

type GetGithubAppResp struct {
	Meta *basedto.Meta  `json:"meta"`
	Data *GithubAppResp `json:"data"`
}

type GithubAppResp struct {
	*settings.BaseSettingResp
	ClientID       string `json:"clientID"`
	ClientSecret   string `json:"clientSecret"`
	Organization   string `json:"organization"`
	CallbackURL    string `json:"callbackURL"`
	WebhookURL     string `json:"webhookURL"`
	WebhookSecret  string `json:"webhookSecret"`
	AppID          int64  `json:"appID"`
	InstallationID int64  `json:"installationID"`
	PrivateKey     string `json:"privateKey"`
	SSOEnabled     bool   `json:"ssoEnabled"`
	SecretMasked   bool   `json:"secretMasked,omitempty"`
}

func (resp *GithubAppResp) CopyClientSecret(field entity.EncryptedField) error {
	resp.ClientSecret = field.String()
	return nil
}

func (resp *GithubAppResp) CopyPrivateKey(field entity.EncryptedField) error {
	resp.PrivateKey = field.String()
	return nil
}

type GithubAppTransformInput struct {
	RefObjects      *entity.RefObjects
	BaseCallbackURL string
}

func TransformGithubApp(
	setting *entity.Setting,
	input *GithubAppTransformInput,
) (resp *GithubAppResp, err error) {
	config := setting.MustAsGithubApp()
	if err = copier.Copy(&resp, config); err != nil {
		return nil, apperrors.Wrap(err)
	}

	resp.BaseSettingResp, err = settings.TransformSettingBase(setting)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	// Recalculate callbackURL for the github-app as it depends on the actual server address
	resp.CallbackURL = input.BaseCallbackURL + "/" + setting.ID
	resp.SecretMasked = config.ClientSecret.IsEncrypted() || resp.Inherited
	if resp.SecretMasked {
		resp.ClientSecret = maskedSecret
		resp.WebhookSecret = maskedSecret
		resp.PrivateKey = maskedSecret
	}

	return resp, nil
}
