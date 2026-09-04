package githubappdto

import (
	"strings"

	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
)

type CreateGithubAppReq struct {
	settings.CreateSettingReq
	*GithubAppBaseReq
}

type GithubAppBaseReq struct {
	Name             string `json:"name"`
	ClientID         string `json:"clientId"`
	ClientSecret     string `json:"clientSecret"`
	Organization     string `json:"organization"`
	GhAppID          int64  `json:"appId"`
	GhInstallationID int64  `json:"installationId"`
	PrivateKey       string `json:"privateKey"`
	SSOEnabled       bool   `json:"ssoEnabled"`
}

func (req *GithubAppBaseReq) ToEntity() *entity.GithubApp {
	return &entity.GithubApp{
		ClientID:       req.ClientID,
		ClientSecret:   entity.NewEncryptedField(req.ClientSecret),
		Organization:   req.Organization,
		AppID:          req.GhAppID,
		InstallationID: req.GhInstallationID,
		PrivateKey:     entity.NewEncryptedField(req.PrivateKey),
		SSOEnabled:     req.SSOEnabled,
	}
}

// KeepMaskedSecrets restores the stored values for the secrets the request only
// carries as the masked placeholder the GET response substitutes for them.
func (req *GithubAppBaseReq) KeepMaskedSecrets(app, current *entity.GithubApp) {
	if current == nil {
		return
	}
	if req.ClientSecret == maskedSecret {
		app.ClientSecret = current.ClientSecret
	}
	if req.PrivateKey == maskedSecret {
		app.PrivateKey = current.PrivateKey
	}
}

func (req *GithubAppBaseReq) modifyRequest() error {
	req.Name = strings.TrimSpace(req.Name)
	req.ClientID = strings.TrimSpace(req.ClientID)
	req.ClientSecret = strings.TrimSpace(req.ClientSecret)
	req.Organization = strings.TrimSpace(req.Organization)
	req.PrivateKey = strings.TrimSpace(req.PrivateKey)
	return nil
}

func (req *GithubAppBaseReq) validate(_ string) []vld.Validator {
	// TODO: add the remaining validation
	var res []vld.Validator
	res = append(res, basedto.ValidatePlainSecret(&req.ClientSecret, "clientSecret")...)
	res = append(res, basedto.ValidatePlainSecret(&req.PrivateKey, "privateKey")...)
	return res
}

func NewCreateGithubAppReq() *CreateGithubAppReq {
	return &CreateGithubAppReq{}
}

func (req *CreateGithubAppReq) ModifyRequest() error {
	return req.modifyRequest()
}

// Validate implements interface basedto.ReqValidator
func (req *CreateGithubAppReq) Validate() hperrors.ValidationErrors {
	validators := make([]vld.Validator, 0, 10) //nolint:mnd
	validators = append(validators, req.CreateSettingReq.Validate()...)
	validators = append(validators, req.validate("")...)
	return hperrors.NewValidationErrors(vld.Validate(validators...))
}

type CreateGithubAppResp struct {
	Meta *basedto.Meta          `json:"meta"`
	Data *GithubAppCreationResp `json:"data"`
}

type GithubAppCreationResp struct {
	ID          string `json:"id"`
	CallbackURL string `json:"callbackURL"`
}
