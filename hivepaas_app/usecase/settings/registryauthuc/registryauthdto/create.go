package registryauthdto

import (
	"strings"

	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
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

// SecretFields lists the request's secret values in one place, so the paths that
// care about them do not each restate the list.
func (req *RegistryAuthBaseReq) SecretFields() []basedto.SecretField {
	return []basedto.SecretField{
		{Path: "password", Value: &req.Password},
	}
}

// KeepMaskedSecrets restores the stored values for the secrets the request only
// carries as the masked placeholder the GET response substitutes for them.
func (req *RegistryAuthBaseReq) KeepMaskedSecrets(regAuth, current *entity.RegistryAuth) {
	if current == nil {
		return
	}
	if basedto.IsMaskedSecret(req.Password) {
		regAuth.Password = current.Password
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
	res = append(res, basedto.ValidatePlainSecret(&req.Password, field+"password")...)
	return res
}

func NewCreateRegistryAuthReq() *CreateRegistryAuthReq {
	return &CreateRegistryAuthReq{}
}

func (req *CreateRegistryAuthReq) ModifyRequest() error {
	return req.modifyRequest()
}

// Validate implements interface basedto.ReqValidator
func (req *CreateRegistryAuthReq) Validate() hperrors.ValidationErrors {
	validators := make([]vld.Validator, 0, 10) //nolint:mnd
	validators = append(validators, req.CreateSettingReq.Validate()...)
	validators = append(validators, req.validate("")...)
	// Creation has no stored value to fall back on, so the placeholder is not a
	// meaningful input here the way it is on update.
	validators = append(validators, basedto.ValidateNoMaskedSecrets(req.SecretFields())...)
	return hperrors.NewValidationErrors(vld.Validate(validators...))
}

type CreateRegistryAuthResp struct {
	Meta *basedto.Meta         `json:"meta"`
	Data *basedto.ObjectIDResp `json:"data"`
}
