package storagesettingsdto

import (
	"github.com/moby/moby/api/types/volume"
	vld "github.com/tiendc/go-validator"
	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/apphelper"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/copier"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/unit"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
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
	Enabled         bool     `json:"enabled,omitempty"`
	BaseDirs        []string `json:"baseDirs"`
	SubpathTemplate string   `json:"subpathTemplate"`
	SubpathRequired string   `json:"subpathRequired,omitempty"` // computed field
}

type StorageVolumeSettingsResp struct {
	Enabled         bool                         `json:"enabled,omitempty"`
	Volumes         basedto.NamedObjectSliceResp `json:"volumes"`
	SubpathTemplate string                       `json:"subpathTemplate"`
	SubpathRequired string                       `json:"subpathRequired,omitempty"` // computed field
}

type StorageClusterVolumeSettingsResp struct {
	Enabled         bool                         `json:"enabled,omitempty"`
	Volumes         basedto.NamedObjectSliceResp `json:"volumes"`
	SubpathTemplate string                       `json:"subpathTemplate"`
	SubpathRequired string                       `json:"subpathRequired,omitempty"` // computed field
}

type StorageTmpfsSettingsResp struct {
	Enabled bool          `json:"enabled,omitempty"`
	MaxSize unit.DataSize `json:"maxSize"`
}

type StorageSettingsTransformInput struct {
	Project *entity.Project
	App     *entity.App
	Setting *entity.Setting
	Volumes []*volume.Volume
}

func TransformStorageSettings(
	input *StorageSettingsTransformInput,
) (resp *StorageSettingsResp, err error) {
	config := input.Setting.MustAsStorageSettings()
	if err = copier.Copy(&resp, config); err != nil {
		return nil, apperrors.New(err)
	}

	if resp.BindSettings != nil {
		// Compute subpath required for scope app
		if input.Project != nil && input.App != nil {
			resp.BindSettings.SubpathRequired = apphelper.CalcMountSubpath(input.Project, input.App,
				resp.BindSettings.SubpathTemplate)
		}
	}

	if resp.VolumeSettings != nil {
		// Docker named volume, the name is the same as the id
		for _, vol := range resp.VolumeSettings.Volumes {
			vol.Name = vol.ID
		}

		// Compute subpath required for scope app
		if input.Project != nil && input.App != nil {
			resp.VolumeSettings.SubpathRequired = apphelper.CalcMountSubpath(input.Project, input.App,
				resp.VolumeSettings.SubpathTemplate)
		}
	}

	if resp.ClusterVolumeSettings != nil {
		// Docker cluster volumes have both IDs and names
		for _, vol := range resp.ClusterVolumeSettings.Volumes {
			v, _ := gofn.Find(input.Volumes, func(v *volume.Volume) bool {
				return v.ClusterVolume != nil && v.ClusterVolume.ID == vol.ID
			})
			if v != nil {
				vol.Name = v.Name
			} else {
				vol.Name = "<missing-volume>"
			}
		}

		// Compute subpath required for scope app
		if input.Project != nil && input.App != nil {
			resp.ClusterVolumeSettings.SubpathRequired = apphelper.CalcMountSubpath(input.Project, input.App,
				resp.ClusterVolumeSettings.SubpathTemplate)
		}
	}

	resp.BaseSettingResp, err = settings.TransformSettingBase(input.Setting)
	if err != nil {
		return nil, apperrors.New(err)
	}

	return resp, nil
}
