package storagesettingsuc

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/storagesettingsuc/storagesettingsdto"
)

func (uc *UC) GetStorageSettings(
	ctx context.Context,
	auth *basedto.Auth,
	req *storagesettingsdto.GetStorageSettingsReq,
) (*storagesettingsdto.GetStorageSettingsResp, error) {
	req.Type = currentSettingType
	resp, err := uc.GetUniqueSetting(ctx, auth, &req.GetUniqueSettingReq, &settings.GetUniqueSettingData{})
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	input := &storagesettingsdto.StorageSettingsTransformInput{
		Setting: resp.Data,
	}

	storageSetting := resp.Data.MustAsStorageSettings()

	// Load reference cluster volumes as their IDs are different from their names
	if storageSetting.ClusterVolumeSettings != nil && len(storageSetting.ClusterVolumeSettings.Volumes) > 0 {
		volResp, err := uc.dockerManager.VolumeListByIDs(ctx,
			storageSetting.ClusterVolumeSettings.Volumes.ToIDStringSlice())
		if err != nil {
			return nil, apperrors.Wrap(err)
		}
		for i := range volResp.Items {
			input.Volumes = append(input.Volumes, &volResp.Items[i])
		}
	}

	respData, err := storagesettingsdto.TransformStorageSettings(input)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	return &storagesettingsdto.GetStorageSettingsResp{
		Data: respData,
	}, nil
}
