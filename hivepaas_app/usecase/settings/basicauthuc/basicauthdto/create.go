package basicauthdto

import (
	"strings"

	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
)

const (
	usernameMaxLen = 100
	passwordMaxLen = 100
)

type CreateBasicAuthReq struct {
	settings.CreateSettingReq
	*BasicAuthBaseReq
}

type BasicAuthBaseReq struct {
	Name     string `json:"name"`
	Username string `json:"username"`
	Password string `json:"password"`
}

func (req *BasicAuthBaseReq) ToEntity() *entity.BasicAuth {
	return &entity.BasicAuth{
		Username: req.Username,
		Password: entity.NewEncryptedField(req.Password),
	}
}

func (req *BasicAuthBaseReq) modifyRequest() error {
	req.Name = strings.TrimSpace(req.Name)
	req.Username = strings.TrimSpace(req.Username)
	return nil
}

func (req *BasicAuthBaseReq) validate(field string) (res []vld.Validator) {
	if field != "" {
		field += "."
	}
	res = append(res, basedto.ValidateStr(&req.Name, true, 1, base.SettingNameMaxLen, field+"name")...)
	res = append(res, basedto.ValidateStr(&req.Username, true, 1, usernameMaxLen, field+"username")...)
	res = append(res, basedto.ValidateStr(&req.Password, true, 1, passwordMaxLen, field+"password")...)
	return res
}

func NewCreateBasicAuthReq() *CreateBasicAuthReq {
	return &CreateBasicAuthReq{}
}

func (req *CreateBasicAuthReq) ModifyRequest() error {
	return req.modifyRequest()
}

// Validate implements interface basedto.ReqValidator
func (req *CreateBasicAuthReq) Validate() apperrors.ValidationErrors {
	var validators []vld.Validator
	validators = append(validators, req.validate("")...)
	return apperrors.NewValidationErrors(vld.Validate(validators...))
}

type CreateBasicAuthResp struct {
	Meta *basedto.Meta         `json:"meta"`
	Data *basedto.ObjectIDResp `json:"data"`
}
