package secretdto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
)

const (
	secretValueMaxLen = 500 * 1024 // 500Kb
)

type CreateSecretReq struct {
	settings.CreateSettingReq
	*SecretBaseReq
}

type SecretBaseReq struct {
	Key    string `json:"key"`
	Value  string `json:"value"`
	Base64 bool   `json:"base64"`
}

func (req *SecretBaseReq) ToEntity() *entity.Secret {
	return &entity.Secret{
		Key:    req.Key,
		Value:  entity.NewEncryptedField(req.Value),
		Base64: req.Base64,
	}
}

func (req *SecretBaseReq) validate(valueRequired bool, field string) (res []vld.Validator) {
	if field != "" {
		field += "."
	}
	res = append(res, basedto.ValidateEnvName(&req.Key, true, field+"key")...)
	if req.Base64 {
		res = append(res, basedto.ValidateStrBase64(&req.Value, valueRequired, 1,
			secretValueMaxLen, field+"value")...)
		res = append(res, basedto.ValidatePlainSecret(&req.Value, field+"value")...)
	} else {
		res = append(res, basedto.ValidateStr(&req.Value, valueRequired, 1,
			secretValueMaxLen, field+"value")...)
		res = append(res, basedto.ValidatePlainSecret(&req.Value, field+"value")...)
	}
	return res
}

func NewCreateSecretReq() *CreateSecretReq {
	return &CreateSecretReq{}
}

// Validate implements interface basedto.ReqValidator
func (req *CreateSecretReq) Validate() hperrors.ValidationErrors {
	validators := make([]vld.Validator, 0, 10) //nolint:mnd
	validators = append(validators, req.CreateSettingReq.Validate()...)
	validators = append(validators, req.validate(true, "")...)
	return hperrors.NewValidationErrors(vld.Validate(validators...))
}

type CreateSecretResp struct {
	Meta *basedto.Meta         `json:"meta"`
	Data *basedto.ObjectIDResp `json:"data"`
}
