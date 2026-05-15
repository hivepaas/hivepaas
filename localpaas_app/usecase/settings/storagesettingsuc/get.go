package storagesettingsuc

import (
	"context"
	"errors"

	"github.com/moby/moby/api/types/volume"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings/storagesettingsuc/storagesettingsdto"
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

	storageSetting := resp.Data.MustAsStorageSettings()
	// Load ref cluster volumes as their IDs are different from their names
	var volumes []*volume.Volume
	//nolint:nestif
	if storageSetting.ClusterVolumeSettings != nil && len(storageSetting.ClusterVolumeSettings.Volumes) > 0 {
		if len(storageSetting.ClusterVolumeSettings.Volumes) == 1 {
			inspect, err := uc.dockerManager.VolumeInspect(ctx, storageSetting.ClusterVolumeSettings.Volumes[0].ID)
			if err != nil && !errors.Is(err, apperrors.ErrNotFound) {
				return nil, apperrors.Wrap(err)
			}
			if inspect != nil {
				volumes = append(volumes, &inspect.Volume)
			}
		} else {
			volResp, err := uc.dockerManager.VolumeList(ctx)
			if err != nil {
				return nil, apperrors.Wrap(err)
			}
			for i := range volResp.Items {
				volumes = append(volumes, &volResp.Items[i])
			}
		}
	}

	respData, err := storagesettingsdto.TransformStorageSettings(resp.Data, volumes)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	return &storagesettingsdto.GetStorageSettingsResp{
		Data: respData,
	}, nil
}
