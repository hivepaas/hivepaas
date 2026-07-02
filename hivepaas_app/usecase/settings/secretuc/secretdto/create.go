package secretdto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/fileutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
)

const (
	secretValueMaxLen = 500 * 1024 // 500Kb
	fileNameMaxLen    = 200
)

type CreateSecretReq struct {
	settings.CreateSettingReq
	*SecretBaseReq
}

type SecretBaseReq struct {
	Key      string             `json:"key"`
	Value    string             `json:"value"`
	Base64   bool               `json:"base64"`
	SwarmRef *SwarmSecretRefReq `json:"swarmRef"`
}

func (req *SecretBaseReq) ToEntity() *entity.Secret {
	return &entity.Secret{
		Key:      req.Key,
		Value:    entity.NewEncryptedField(req.Value),
		Base64:   req.Base64,
		SwarmRef: req.SwarmRef.ToEntity(),
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
	} else {
		res = append(res, basedto.ValidateStr(&req.Value, valueRequired, 1,
			secretValueMaxLen, field+"value")...)
	}
	res = append(res, req.SwarmRef.validate(field+"swarmRef")...)
	return res
}

type SwarmSecretRefReq struct {
	File *SwarmRefFileTargetReq `json:"file"`
}

func (req *SwarmSecretRefReq) ToEntity() *entity.SwarmSecretRef {
	if req == nil {
		return nil
	}
	return &entity.SwarmSecretRef{
		File: req.File.ToEntity(),
	}
}

func (req *SwarmSecretRefReq) validate(field string) (res []vld.Validator) {
	if req == nil {
		return nil
	}
	if field != "" {
		field += "."
	}
	res = append(res, req.File.validate(field+"file")...)
	return res
}

type SwarmRefFileTargetReq struct {
	Name string            `json:"name"`
	UID  string            `json:"uid"`
	GID  string            `json:"gid"`
	Mode fileutil.FileMode `json:"mode"`
}

func (req *SwarmRefFileTargetReq) ToEntity() *entity.SwarmRefFileTarget {
	if req == nil {
		return nil
	}
	return &entity.SwarmRefFileTarget{
		Name: req.Name,
		UID:  req.UID,
		GID:  req.GID,
		Mode: req.Mode,
	}
}

func (req *SwarmRefFileTargetReq) validate(field string) (res []vld.Validator) {
	if req == nil {
		return nil
	}
	if field != "" {
		field += "."
	}
	res = append(res, basedto.ValidateStr(&req.Name, false, 1, fileNameMaxLen, field+"name")...)
	return res
}

func NewCreateSecretReq() *CreateSecretReq {
	return &CreateSecretReq{}
}

// Validate implements interface basedto.ReqValidator
func (req *CreateSecretReq) Validate() apperrors.ValidationErrors {
	var validators []vld.Validator
	validators = append(validators, req.validate(true, "")...)
	return apperrors.NewValidationErrors(vld.Validate(validators...))
}

type CreateSecretResp struct {
	Meta *basedto.Meta         `json:"meta"`
	Data *basedto.ObjectIDResp `json:"data"`
}
