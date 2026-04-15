package storagesettingsdto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/entity"
	"github.com/localpaas/localpaas/localpaas_app/pkg/unit"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings"
)

type UpdateUniqueStorageSettingsReq struct {
	settings.UpdateUniqueSettingReq
	*StorageSettingsBaseReq
}

type StorageSettingsBaseReq struct {
	BindSettings   *StorageBindSettingsReq   `json:"bindSettings"`
	VolumeSettings *StorageVolumeSettingsReq `json:"volumeSettings"`
	TmpfsSettings  *StorageTmpfsSettingsReq  `json:"tmpfsSettings"`
}

func (req *StorageSettingsBaseReq) ToEntity() *entity.StorageSettings {
	return &entity.StorageSettings{
		BindSettings:   req.BindSettings.ToEntity(),
		VolumeSettings: req.VolumeSettings.ToEntity(),
		TmpfsSettings:  req.TmpfsSettings.ToEntity(),
	}
}

func (req *StorageSettingsBaseReq) validate(field string) (res []vld.Validator) {
	if field != "" {
		field += "."
	}
	res = append(res, req.BindSettings.validate(field+"bindSettings")...)
	res = append(res, req.BindSettings.validate(field+"volumeSettings")...)
	res = append(res, req.BindSettings.validate(field+"tmpfsSettings")...)
	return res
}

type StorageBindSettingsReq struct {
	AllowAny            bool     `json:"allowAny,omitempty"`
	BaseDirs            []string `json:"baseDirs"`
	AppsMustUseSubPaths bool     `json:"appsMustUseSubPaths"`
}

func (req *StorageBindSettingsReq) ToEntity() *entity.StorageBindSettings {
	if req == nil {
		return nil
	}
	return &entity.StorageBindSettings{
		AllowAny:            req.AllowAny,
		BaseDirs:            req.BaseDirs,
		AppsMustUseSubPaths: req.AppsMustUseSubPaths,
	}
}

// nolint
func (req *StorageBindSettingsReq) validate(field string) (res []vld.Validator) {
	if field != "" {
		field += "."
	}
	// TODO: add validation
	return res
}

type StorageVolumeSettingsReq struct {
	AllowAny            bool     `json:"allowAny,omitempty"`
	Volumes             []string `json:"volumes"`
	AppsMustUseSubPaths bool     `json:"appsMustUseSubPaths"`
}

func (req *StorageVolumeSettingsReq) ToEntity() *entity.StorageVolumeSettings {
	if req == nil {
		return nil
	}
	return &entity.StorageVolumeSettings{
		AllowAny:            req.AllowAny,
		Volumes:             req.Volumes,
		AppsMustUseSubPaths: req.AppsMustUseSubPaths,
	}
}

// nolint
func (req *StorageVolumeSettingsReq) validate(field string) (res []vld.Validator) {
	if field != "" {
		field += "."
	}
	// TODO: add validation
	return res
}

type StorageTmpfsSettingsReq struct {
	MaxSize unit.DataSize `json:"maxSize"`
}

func (req *StorageTmpfsSettingsReq) ToEntity() *entity.StorageTmpfsSettings {
	if req == nil {
		return nil
	}
	return &entity.StorageTmpfsSettings{
		MaxSize: req.MaxSize,
	}
}

// nolint
func (req *StorageTmpfsSettingsReq) validate(field string) (res []vld.Validator) {
	if field != "" {
		field += "."
	}
	// TODO: add validation
	return res
}

func NewUpdateUniqueStorageSettingsReq() *UpdateUniqueStorageSettingsReq {
	return &UpdateUniqueStorageSettingsReq{}
}

// Validate implements interface basedto.ReqValidator
func (req *UpdateUniqueStorageSettingsReq) Validate() apperrors.ValidationErrors {
	var validators []vld.Validator
	validators = append(validators, req.validate("")...)
	return apperrors.NewValidationErrors(vld.Validate(validators...))
}

type UpdateUniqueStorageSettingsResp struct {
	Meta *basedto.Meta `json:"meta"`
}
