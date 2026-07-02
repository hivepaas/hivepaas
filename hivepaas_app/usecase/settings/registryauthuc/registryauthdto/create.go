package registryauthdto

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
	addressMaxLen  = 200
	usernameMaxLen = 100
	passwordMaxLen = 100
)

type CreateRegistryAuthReq struct {
	settings.CreateSettingReq
	*RegistryAuthBaseReq
}

type RegistryAuthBaseReq struct {
	Name     string `json:"name"`
	Address  string `json:"address"`
	Username string `json:"username"`
	Password string `json:"password"`
	Readonly bool   `json:"readonly"`
}

func (req *RegistryAuthBaseReq) ToEntity() *entity.RegistryAuth {
	return &entity.RegistryAuth{
		Username: req.Username,
		Password: entity.NewEncryptedField(req.Password),
		Address:  req.Address,
		Readonly: req.Readonly,
	}
}

func (req *RegistryAuthBaseReq) modifyRequest() error {
	req.Name = strings.TrimSpace(req.Name)
	req.Address = strings.ToLower(strings.TrimSpace(req.Address))
	req.Username = strings.TrimSpace(req.Username)
	return nil
}

func (req *RegistryAuthBaseReq) validate(field string) (res []vld.Validator) {
	if field != "" {
		field += "."
	}
	res = append(res, basedto.ValidateStr(&req.Name, true, 1, base.SettingNameMaxLen, field+"name")...)
	res = append(res, basedto.ValidateStr(&req.Address, true, 1, addressMaxLen, field+"address")...)
	res = append(res, basedto.ValidateStr(&req.Username, true, 1, usernameMaxLen, field+"username")...)
	res = append(res, basedto.ValidateStr(&req.Password, true, 1, passwordMaxLen, field+"password")...)
	return res
}

func NewCreateRegistryAuthReq() *CreateRegistryAuthReq {
	return &CreateRegistryAuthReq{}
}

func (req *CreateRegistryAuthReq) ModifyRequest() error {
	return req.modifyRequest()
}

// Validate implements interface basedto.ReqValidator
func (req *CreateRegistryAuthReq) Validate() apperrors.ValidationErrors {
	var validators []vld.Validator
	validators = append(validators, req.validate("")...)
	return apperrors.NewValidationErrors(vld.Validate(validators...))
}

type CreateRegistryAuthResp struct {
	Meta *basedto.Meta         `json:"meta"`
	Data *basedto.ObjectIDResp `json:"data"`
}
