package secretdto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/entity"
	"github.com/localpaas/localpaas/localpaas_app/pkg/osutil"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings"
)

const (
	secretKeyMaxLen   = 200
	secretValueMaxLen = 10 * 1024 * 1024 // 10MB
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
	res = append(res, basedto.ValidateStr(&req.Key, true, 1, secretKeyMaxLen, field+"key")...)
	res = append(res, basedto.ValidateStr(&req.Value, valueRequired, 1, secretValueMaxLen, field+"value")...)
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
	Name string          `json:"name"`
	UID  string          `json:"uid"`
	GID  string          `json:"gid"`
	Mode osutil.FileMode `json:"mode"`
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
	res = append(res, basedto.ValidateStr(&req.Name, false, 1, secretKeyMaxLen, field+"name")...)
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
