package accesstokendto

import (
	"strings"
	"time"

	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
)

const (
	tokenMaxLen = 512
	urlMaxLen   = 512
)

type CreateAccessTokenReq struct {
	settings.CreateSettingReq
	*AccessTokenBaseReq
}

type AccessTokenBaseReq struct {
	Kind     base.AccessTokenKind `json:"kind"`
	Name     string               `json:"name"`
	User     string               `json:"user"`
	Token    string               `json:"token"`
	BaseURL  string               `json:"baseURL"`
	ExpireAt time.Time            `json:"expireAt"`
}

func (req *AccessTokenBaseReq) ToEntity() *entity.AccessToken {
	return &entity.AccessToken{
		User:    req.User,
		Token:   entity.NewEncryptedField(req.Token),
		BaseURL: req.BaseURL,
	}
}

func (req *AccessTokenBaseReq) modifyRequest() error {
	req.Name = strings.TrimSpace(req.Name)
	req.User = strings.TrimSpace(req.User)
	req.Token = strings.TrimSpace(req.Token)
	req.BaseURL = strings.TrimSpace(req.BaseURL)
	return nil
}

func (req *AccessTokenBaseReq) validate(field string) (res []vld.Validator) {
	if field != "" {
		field += "."
	}
	res = append(res, basedto.ValidateStrIn(&req.Kind, true, base.AllAccessTokenKinds, field+"kind")...)
	res = append(res, basedto.ValidateStr(&req.Name, true, 1, base.SettingNameMaxLen, field+"name")...)
	res = append(res, basedto.ValidateStr(&req.Token, true, 1, tokenMaxLen, field+"token")...)
	res = append(res, basedto.ValidateStr(&req.User, false, 1, base.SettingNameMaxLen, field+"user")...)
	res = append(res, basedto.ValidateStr(&req.BaseURL, false, 1, urlMaxLen, field+"baseURL")...)
	return res
}

func NewCreateAccessTokenReq() *CreateAccessTokenReq {
	return &CreateAccessTokenReq{}
}

func (req *CreateAccessTokenReq) ModifyRequest() error {
	return req.modifyRequest()
}

// Validate implements interface basedto.ReqValidator
func (req *CreateAccessTokenReq) Validate() apperrors.ValidationErrors {
	var validators []vld.Validator
	validators = append(validators, req.validate("")...)
	return apperrors.NewValidationErrors(vld.Validate(validators...))
}

type CreateAccessTokenResp struct {
	Meta *basedto.Meta         `json:"meta"`
	Data *basedto.ObjectIDResp `json:"data"`
}
