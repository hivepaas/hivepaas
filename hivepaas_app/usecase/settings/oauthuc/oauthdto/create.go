package oauthdto

import (
	"strings"
	"unicode"

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
	req.Scopes = strings.FieldsFunc(strings.Join(req.Scopes, ","), func(r rune) bool {
		return r == ',' || unicode.IsSpace(r)
	})
	if len(req.Scopes) == 0 {
		switch req.Kind {
		case base.OAuthKindGitea:
			req.Scopes = strings.Fields(base.OAuthScopeDefaultGitea)
		case base.OAuthKindGithub:
			req.Scopes = strings.Fields(base.OAuthScopeDefaultGithub)
		case base.OAuthKindGithubApp:
		case base.OAuthKindGitlab:
			req.Scopes = strings.Fields(base.OAuthScopeDefaultGitlab)
		case base.OAuthKindGoogle:
			req.Scopes = strings.Fields(base.OAuthScopeDefaultGoogle)
		case base.OAuthKindMicrosoftOnline:
			req.Scopes = strings.Fields(base.OAuthScopeDefaultMicrosoftOnline)
		case base.OAuthKindOpenIDConnect:
			req.Scopes = strings.Fields(base.OAuthScopeDefaultOpenIDConnect)
		}
	}
	return nil
}

func (req *OAuthBaseReq) validate(field string) (res []vld.Validator) {
	if field != "" {
		field += "."
	}
	res = append(res, basedto.ValidateStr(&req.Name, true, base.SettingNameMinLen,
		base.SettingNameMaxLen, field+"name")...)
	res = append(res, basedto.ValidateStr(&req.ClientID, true, base.IDMinLen,
		base.IDMaxLen, field+"clientId")...)
	res = append(res, basedto.ValidateStr(&req.ClientSecret, true, base.SecretMinLen,
		base.SecretMaxLen, field+"clientSecret")...)
	res = append(res, basedto.ValidateStr(&req.Organization, false, base.IDMinLen,
		base.IDMaxLen, field+"organization")...)
	res = append(res, basedto.ValidateStr(&req.AuthURL, false, base.URLMinLen,
		base.URLMaxLen, field+"authURL")...)
	res = append(res, basedto.ValidateStr(&req.TokenURL, false, base.URLMinLen,
		base.URLMaxLen, field+"tokenURL")...)
	res = append(res, basedto.ValidateStr(&req.ProfileURL, false, base.URLMinLen,
		base.URLMaxLen, field+"profileURL")...)
	res = append(res, basedto.ValidateStr(&req.AutoDiscoveryURL, false, base.URLMinLen,
		base.URLMaxLen, field+"autoDiscoveryURL")...)
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
