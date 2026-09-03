package githubappdto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/copier"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
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

func (req *GetGithubAppReq) Validate() hperrors.ValidationErrors {
	validators := make([]vld.Validator, 0, 10) //nolint:mnd
	validators = append(validators, req.GetSettingReq.Validate()...)
	return hperrors.NewValidationErrors(vld.Validate(validators...))
}

type GetGithubAppResp struct {
	Meta *basedto.Meta  `json:"meta"`
	Data *GithubAppResp `json:"data"`
}

type GithubAppResp struct {
	*settings.BaseSettingResp
	ClientID     string `json:"clientId"`
	ClientSecret string `json:"clientSecret"`
	Organization string `json:"organization"`
	Slug         string `json:"slug,omitempty"`
	OwnerLogin   string `json:"ownerLogin,omitempty"`
	OwnerType    string `json:"ownerType,omitempty"`
	// SettingsURL / InstallationsURL / PublicURL are the github.com pages of the
	// app. They are built from the slug, never from the name.
	SettingsURL      string `json:"settingsURL,omitempty"`
	InstallationsURL string `json:"installationsURL,omitempty"`
	PublicURL        string `json:"publicURL,omitempty"`
	CallbackURL      string `json:"callbackURL"`
	WebhookURL       string `json:"webhookURL"`
	WebhookSecret    string `json:"webhookSecret"`
	AppID            int64  `json:"appId"`
	InstallationID   int64  `json:"installationId"`
	PrivateKey       string `json:"privateKey"`
	SSOEnabled       bool   `json:"ssoEnabled"`
	SecretMasked     bool   `json:"secretMasked,omitempty"`
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
		return nil, hperrors.Wrap(err)
	}

	resp.BaseSettingResp, err = settings.TransformSettingBase(setting)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	// Recalculate callbackURL for the github-app as it depends on the actual server address
	resp.CallbackURL = input.BaseCallbackURL + "/" + setting.ID
	resp.SettingsURL = config.SettingsURL()
	resp.InstallationsURL = config.InstallationsURL()
	resp.PublicURL = config.PublicURL()
	resp.SecretMasked = config.ClientSecret.IsEncrypted() || resp.Inherited
	if resp.SecretMasked {
		resp.ClientSecret = maskedSecret
		resp.WebhookSecret = maskedSecret
		resp.PrivateKey = maskedSecret
	}

	return resp, nil
}
