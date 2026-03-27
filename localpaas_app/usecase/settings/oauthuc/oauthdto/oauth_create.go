package oauthdto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/base"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/entity"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings"
)

type CreateOAuthReq struct {
	settings.CreateSettingReq
	*OAuthBaseReq
}

type OAuthBaseReq struct {
	Kind         base.OAuthKind `json:"kind"`
	Name         string         `json:"name"`
	ClientID     string         `json:"clientID"`
	ClientSecret string         `json:"clientSecret"`
	Organization string         `json:"organization"`
	AuthURL      string         `json:"authURL"`
	TokenURL     string         `json:"tokenURL"`
	ProfileURL   string         `json:"profileURL"`
	Scopes       []string       `json:"scopes"`
}

func (req *OAuthBaseReq) ToEntity() *entity.OAuth {
	return &entity.OAuth{
		ClientID:     req.ClientID,
		ClientSecret: entity.NewEncryptedField(req.ClientSecret),
		Organization: req.Organization,
		AuthURL:      req.AuthURL,
		TokenURL:     req.TokenURL,
		ProfileURL:   req.ProfileURL,
		Scopes:       req.Scopes,
	}
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
