package githubappdto

import (
	"strings"

	vld "github.com/tiendc/go-validator"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/entity"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings"
)

type CreateGithubAppReq struct {
	settings.CreateSettingReq
	*GithubAppBaseReq
}

type GithubAppBaseReq struct {
	Name             string `json:"name"`
	ClientID         string `json:"clientID"`
	ClientSecret     string `json:"clientSecret"`
	Organization     string `json:"organization"`
	GhAppID          int64  `json:"appID"`
	GhInstallationID int64  `json:"installationID"`
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

func (req *GithubAppBaseReq) modifyRequest() error {
	req.Name = strings.TrimSpace(req.Name)
	req.ClientID = strings.TrimSpace(req.ClientID)
	req.ClientSecret = strings.TrimSpace(req.ClientSecret)
	req.Organization = strings.TrimSpace(req.Organization)
	req.PrivateKey = strings.TrimSpace(req.PrivateKey)
	return nil
}

func (req *GithubAppBaseReq) validate(_ string) []vld.Validator {
	// TODO: add validation
	return nil
}

func NewCreateGithubAppReq() *CreateGithubAppReq {
	return &CreateGithubAppReq{}
}

func (req *CreateGithubAppReq) ModifyRequest() error {
	return req.modifyRequest()
}

// Validate implements interface basedto.ReqValidator
func (req *CreateGithubAppReq) Validate() apperrors.ValidationErrors {
	var validators []vld.Validator
	validators = append(validators, req.validate("")...)
	return apperrors.NewValidationErrors(vld.Validate(validators...))
}

type CreateGithubAppResp struct {
	Meta *basedto.Meta          `json:"meta"`
	Data *GithubAppCreationResp `json:"data"`
}

type GithubAppCreationResp struct {
	ID          string `json:"id"`
	CallbackURL string `json:"callbackURL"`
}
