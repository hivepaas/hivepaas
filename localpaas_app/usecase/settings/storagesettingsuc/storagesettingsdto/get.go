package storagesettingsdto

import (
	"github.com/moby/moby/api/types/volume"
	vld "github.com/tiendc/go-validator"
	"github.com/tiendc/gofn"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/entity"
	"github.com/localpaas/localpaas/localpaas_app/pkg/copier"
	"github.com/localpaas/localpaas/localpaas_app/pkg/unit"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings"
)

type GetStorageSettingsReq struct {
	settings.GetUniqueSettingReq
}

func NewGetStorageSettingsReq() *GetStorageSettingsReq {
	return &GetStorageSettingsReq{}
}

func (req *GetStorageSettingsReq) Validate() apperrors.ValidationErrors {
	var validators []vld.Validator
	validators = append(validators, req.GetUniqueSettingReq.Validate()...)
	return apperrors.NewValidationErrors(vld.Validate(validators...))
}

type GetStorageSettingsResp struct {
	Meta *basedto.Meta        `json:"meta"`
	Data *StorageSettingsResp `json:"data"`
}

type StorageSettingsResp struct {
	*settings.BaseSettingResp
	BindSettings          *StorageBindSettingsResp          `json:"bindSettings"`
	VolumeSettings        *StorageVolumeSettingsResp        `json:"volumeSettings"`
	ClusterVolumeSettings *StorageClusterVolumeSettingsResp `json:"clusterVolumeSettings"`
	TmpfsSettings         *StorageTmpfsSettingsResp         `json:"tmpfsSettings"`
}

type StorageBindSettingsResp struct {
	Enabled             bool     `json:"enabled,omitempty"`
	BaseDirs            []string `json:"baseDirs"`
	BaseSubpath         string   `json:"baseSubpath"`
	AppsMustUseSubPaths bool     `json:"appsMustUseSubPaths"`
}

type StorageVolumeSettingsResp struct {
	Enabled             bool                         `json:"enabled,omitempty"`
	Volumes             basedto.NamedObjectSliceResp `json:"volumes"`
	BaseSubpath         string                       `json:"baseSubpath"`
	AppsMustUseSubPaths bool                         `json:"appsMustUseSubPaths"`
}

type StorageClusterVolumeSettingsResp struct {
	Enabled             bool                         `json:"enabled,omitempty"`
	Volumes             basedto.NamedObjectSliceResp `json:"volumes"`
	BaseSubpath         string                       `json:"baseSubpath"`
	AppsMustUseSubPaths bool                         `json:"appsMustUseSubPaths"`
}

type StorageTmpfsSettingsResp struct {
	Enabled bool          `json:"enabled,omitempty"`
	MaxSize unit.DataSize `json:"maxSize"`
}

func TransformStorageSettings(
	setting *entity.Setting,
	volumes []*volume.Volume,
) (resp *StorageSettingsResp, err error) {
	config := setting.MustAsStorageSettings()
	if err = copier.Copy(&resp, config); err != nil {
		return nil, apperrors.Wrap(err)
	}

	// Docker named volume, the name is the same as the id
	if resp.VolumeSettings != nil {
		for _, vol := range resp.VolumeSettings.Volumes {
			vol.Name = vol.ID
		}
	}

	// Docker cluster volumes have both IDs and names
	if resp.ClusterVolumeSettings != nil {
		for _, vol := range resp.ClusterVolumeSettings.Volumes {
			v, _ := gofn.Find(volumes, func(v *volume.Volume) bool {
				return v.ClusterVolume != nil && v.ClusterVolume.ID == vol.ID
			})
			if v != nil {
				vol.Name = v.Name
			} else {
				vol.Name = "<missing-volume>"
			}
		}
	}

	resp.BaseSettingResp, err = settings.TransformSettingBase(setting)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	return resp, nil
}
