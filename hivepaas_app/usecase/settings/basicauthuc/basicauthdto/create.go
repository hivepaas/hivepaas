package basicauthdto

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

// SecretFields lists the request's secret values in one place, so the paths that
// care about them do not each restate the list.
func (req *BasicAuthBaseReq) SecretFields() []basedto.SecretField {
	return []basedto.SecretField{{Path: "password", Value: &req.Password}}
}

// KeepMaskedSecrets restores the stored values for the secrets the request only
// carries as the masked placeholder the GET response substitutes for them.
func (req *BasicAuthBaseReq) KeepMaskedSecrets(basicAuth, current *entity.BasicAuth) {
	if current == nil {
		return
	}
	if basedto.IsMaskedSecret(req.Password) {
		basicAuth.Password = current.Password
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
	res = append(res, basedto.ValidatePlainSecret(&req.Password, field+"password")...)
	return res
}

func NewCreateBasicAuthReq() *CreateBasicAuthReq {
	return &CreateBasicAuthReq{}
}

func (req *CreateBasicAuthReq) ModifyRequest() error {
	return req.modifyRequest()
}

// Validate implements interface basedto.ReqValidator
func (req *CreateBasicAuthReq) Validate() hperrors.ValidationErrors {
	validators := make([]vld.Validator, 0, 10) //nolint:mnd
	validators = append(validators, req.CreateSettingReq.Validate()...)
	validators = append(validators, req.validate("")...)
	// Creation has no stored value to fall back on, so the placeholder is not a
	// meaningful input here the way it is on update.
	validators = append(validators, basedto.ValidateNoMaskedSecrets(req.SecretFields())...)
	return hperrors.NewValidationErrors(vld.Validate(validators...))
}

type CreateBasicAuthResp struct {
	Meta *basedto.Meta         `json:"meta"`
	Data *basedto.ObjectIDResp `json:"data"`
}
