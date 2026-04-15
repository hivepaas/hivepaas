package storagesettingsdto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/entity"
	"github.com/localpaas/localpaas/localpaas_app/pkg/copier"
	"github.com/localpaas/localpaas/localpaas_app/pkg/unit"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings"
)

type GetUniqueStorageSettingsReq struct {
	settings.GetUniqueSettingReq
}

func NewGetUniqueStorageSettingsReq() *GetUniqueStorageSettingsReq {
	return &GetUniqueStorageSettingsReq{}
}

func (req *GetUniqueStorageSettingsReq) Validate() apperrors.ValidationErrors {
	var validators []vld.Validator
	validators = append(validators, req.GetUniqueSettingReq.Validate()...)
	return apperrors.NewValidationErrors(vld.Validate(validators...))
}

type GetUniqueStorageSettingsResp struct {
	Meta *basedto.Meta        `json:"meta"`
	Data *StorageSettingsResp `json:"data"`
}

type StorageSettingsResp struct {
	*settings.BaseSettingResp
	BindSettings   *StorageBindSettingsResp   `json:"bindSettings"`
	VolumeSettings *StorageVolumeSettingsResp `json:"volumeSettings"`
	TmpfsSettings  *StorageTmpfsSettingsResp  `json:"tmpfsSettings"`
}

type StorageBindSettingsResp struct {
	AllowAny            bool     `json:"allowAny,omitempty"`
	BaseDirs            []string `json:"baseDirs"`
	AppsMustUseSubPaths bool     `json:"appsMustUseSubPaths"`
}

type StorageVolumeSettingsResp struct {
	AllowAny            bool     `json:"allowAny,omitempty"`
	Volumes             []string `json:"volumes"`
	AppsMustUseSubPaths bool     `json:"appsMustUseSubPaths"`
}

type StorageTmpfsSettingsResp struct {
	MaxSize unit.DataSize `json:"maxSize"`
}

func TransformStorageSettings(
	setting *entity.Setting,
	_ *entity.RefObjects,
) (resp *StorageSettingsResp, err error) {
	config := setting.MustAsStorageSettings()
	if err = copier.Copy(&resp, config); err != nil {
		return nil, apperrors.Wrap(err)
	}

	resp.BaseSettingResp, err = settings.TransformSettingBase(setting)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	return resp, nil
}
