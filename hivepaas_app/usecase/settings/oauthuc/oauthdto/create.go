package oauthdto

import (
	"strings"

	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
)

type CreateOAuthReq struct {
	settings.CreateSettingReq
	*OAuthBaseReq
}

type OAuthBaseReq struct {
	Kind             base.OAuthKind `json:"kind"`
	Name             string         `json:"name"`
	ClientID         string         `json:"clientId"`
	ClientSecret     string         `json:"clientSecret"`
	Organization     string         `json:"organization"`
	AuthURL          string         `json:"authURL"`
	TokenURL         string         `json:"tokenURL"`
	ProfileURL       string         `json:"profileURL"`
	AutoDiscoveryURL string         `json:"autoDiscoveryURL"`
	Scopes           []string       `json:"scopes"`
}

func (req *OAuthBaseReq) ToEntity() *entity.OAuth {
	return &entity.OAuth{
		ClientID:         req.ClientID,
		ClientSecret:     entity.NewEncryptedField(req.ClientSecret),
		Organization:     req.Organization,
		AuthURL:          req.AuthURL,
		TokenURL:         req.TokenURL,
		ProfileURL:       req.ProfileURL,
		AutoDiscoveryURL: req.AutoDiscoveryURL,
		Scopes:           req.Scopes,
	}
}

func (req *OAuthBaseReq) modifyRequest() error {
	req.ClientID = strings.TrimSpace(req.ClientID)
	req.Organization = strings.TrimSpace(req.Organization)
	req.AuthURL = strings.TrimSpace(req.AuthURL)
	req.TokenURL = strings.TrimSpace(req.TokenURL)
	req.ProfileURL = strings.TrimSpace(req.ProfileURL)
	req.AutoDiscoveryURL = strings.TrimSpace(req.AutoDiscoveryURL)
	for i := range req.Scopes {
		req.Scopes[i] = strings.TrimSpace(req.Scopes[i])
	}
	return nil
}

// nolint
func (req *OAuthBaseReq) validate(field string) (res []vld.Validator) {
	if field != "" {
		field += "."
	}
	// TODO: add validation
	return res
}

func NewCreateOAuthReq() *CreateOAuthReq {
	return &CreateOAuthReq{}
}

func (req *CreateOAuthReq) ModifyRequest() error {
	return req.modifyRequest()
}

// Validate implements interface basedto.ReqValidator
func (req *CreateOAuthReq) Validate() apperrors.ValidationErrors {
	var validators []vld.Validator
	validators = append(validators, req.validate("")...)
	return apperrors.NewValidationErrors(vld.Validate(validators...))
}

type CreateOAuthResp struct {
	Meta *basedto.Meta      `json:"meta"`
	Data *OAuthCreationResp `json:"data"`
}

type OAuthCreationResp struct {
	ID          string `json:"id"`
	CallbackURL string `json:"callbackURL"`
}
