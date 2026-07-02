package configfiledto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/fileutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
)

const (
	configFileNameMaxLen    = 200
	configFileContentMaxLen = 1024 * 1024 // 1MB
)

type CreateConfigFileReq struct {
	settings.CreateSettingReq
	*ConfigFileBaseReq
}

type ConfigFileBaseReq struct {
	Name     string             `json:"name"`
	Content  string             `json:"content"`
	Base64   bool               `json:"base64"`
	SwarmRef *SwarmConfigRefReq `json:"swarmRef"`
}

func (req *ConfigFileBaseReq) ToEntity() *entity.ConfigFile {
	return &entity.ConfigFile{
		Name:     req.Name,
		Content:  req.Content,
		Base64:   req.Base64,
		SwarmRef: req.SwarmRef.ToEntity(),
	}
}

func (req *ConfigFileBaseReq) validate(valueRequired bool, field string) (res []vld.Validator) {
	if req == nil {
		return nil
	}
	if field != "" {
		field += "."
	}
	res = append(res, basedto.ValidateStr(&req.Name, true, 1, configFileNameMaxLen, field+"name")...)
	if req.Base64 {
		res = append(res, basedto.ValidateStrBase64(&req.Content, valueRequired, 1,
			configFileContentMaxLen, field+"content")...)
	} else {
		res = append(res, basedto.ValidateStr(&req.Content, valueRequired, 1,
			configFileContentMaxLen, field+"content")...)
	}
	res = append(res, req.SwarmRef.validate(field+"swarmRef")...)
	return res
}

type SwarmConfigRefReq struct {
	File *SwarmRefFileTargetReq `json:"file"`
}

func (req *SwarmConfigRefReq) ToEntity() *entity.SwarmConfigRef {
	if req == nil {
		return nil
	}
	return &entity.SwarmConfigRef{
		File: req.File.ToEntity(),
	}
}

func (req *SwarmConfigRefReq) validate(field string) (res []vld.Validator) {
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
	res = append(res, basedto.ValidateStr(&req.Name, false, 1, configFileNameMaxLen, field+"name")...)
	return res
}

func NewCreateConfigFileReq() *CreateConfigFileReq {
	return &CreateConfigFileReq{}
}

// Validate implements interface basedto.ReqValidator
func (req *CreateConfigFileReq) Validate() apperrors.ValidationErrors {
	var validators []vld.Validator
	validators = append(validators, req.validate(true, "")...)
	return apperrors.NewValidationErrors(vld.Validate(validators...))
}

type CreateConfigFileResp struct {
	Meta *basedto.Meta         `json:"meta"`
	Data *basedto.ObjectIDResp `json:"data"`
}
