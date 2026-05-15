package storagesettingsdto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/entity"
	"github.com/localpaas/localpaas/localpaas_app/pkg/unit"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings"
)

type UpdateStorageSettingsReq struct {
	settings.UpdateUniqueSettingReq
	*StorageSettingsBaseReq
}

type StorageSettingsBaseReq struct {
	BindSettings          *StorageBindSettingsReq          `json:"bindSettings"`
	VolumeSettings        *StorageVolumeSettingsReq        `json:"volumeSettings"`
	ClusterVolumeSettings *StorageClusterVolumeSettingsReq `json:"clusterVolumeSettings"`
	TmpfsSettings         *StorageTmpfsSettingsReq         `json:"tmpfsSettings"`
}

func (req *StorageSettingsBaseReq) ToEntity() *entity.StorageSettings {
	return &entity.StorageSettings{
		BindSettings:          req.BindSettings.ToEntity(),
		VolumeSettings:        req.VolumeSettings.ToEntity(),
		ClusterVolumeSettings: req.ClusterVolumeSettings.ToEntity(),
		TmpfsSettings:         req.TmpfsSettings.ToEntity(),
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
	Enabled             bool     `json:"enabled"`
	BaseDirs            []string `json:"baseDirs"`
	BaseSubpath         string   `json:"baseSubpath"`
	AppsMustUseSubPaths bool     `json:"appsMustUseSubPaths"`
}

func (req *StorageBindSettingsReq) ToEntity() *entity.StorageBindSettings {
	if req == nil {
		return nil
	}
	return &entity.StorageBindSettings{
		Enabled:             req.Enabled,
		BaseDirs:            req.BaseDirs,
		BaseSubpath:         req.BaseSubpath,
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
	Enabled             bool                     `json:"enabled"`
	Volumes             basedto.ObjectIDSliceReq `json:"volumes"`
	BaseSubpath         string                   `json:"baseSubpath"`
	AppsMustUseSubPaths bool                     `json:"appsMustUseSubPaths"`
}

func (req *StorageVolumeSettingsReq) ToEntity() *entity.StorageVolumeSettings {
	if req == nil {
		return nil
	}
	return &entity.StorageVolumeSettings{
		Enabled:             req.Enabled,
		Volumes:             req.Volumes.ToEntity(),
		BaseSubpath:         req.BaseSubpath,
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

type StorageClusterVolumeSettingsReq struct {
	Enabled             bool                     `json:"enabled"`
	Volumes             basedto.ObjectIDSliceReq `json:"volumes"`
	BaseSubpath         string                   `json:"baseSubpath"`
	AppsMustUseSubPaths bool                     `json:"appsMustUseSubPaths"`
}

func (req *StorageClusterVolumeSettingsReq) ToEntity() *entity.StorageClusterVolumeSettings {
	if req == nil {
		return nil
	}
	return &entity.StorageClusterVolumeSettings{
		Enabled:             req.Enabled,
		Volumes:             req.Volumes.ToEntity(),
		BaseSubpath:         req.BaseSubpath,
		AppsMustUseSubPaths: req.AppsMustUseSubPaths,
	}
}

// nolint
func (req *StorageClusterVolumeSettingsReq) validate(field string) (res []vld.Validator) {
	if field != "" {
		field += "."
	}
	// TODO: add validation
	return res
}

type StorageTmpfsSettingsReq struct {
	Enabled bool          `json:"enabled"`
	MaxSize unit.DataSize `json:"maxSize"`
}

func (req *StorageTmpfsSettingsReq) ToEntity() *entity.StorageTmpfsSettings {
	if req == nil {
		return nil
	}
	return &entity.StorageTmpfsSettings{
		Enabled: req.Enabled,
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

func NewUpdateStorageSettingsReq() *UpdateStorageSettingsReq {
	return &UpdateStorageSettingsReq{}
}

// Validate implements interface basedto.ReqValidator
func (req *UpdateStorageSettingsReq) Validate() apperrors.ValidationErrors {
	var validators []vld.Validator
	validators = append(validators, req.validate("")...)
	return apperrors.NewValidationErrors(vld.Validate(validators...))
}

type UpdateStorageSettingsResp struct {
	Meta *basedto.Meta `json:"meta"`
}
