package backuprepocleanupuc

import (
	"context"
	"errors"

	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/systemsettings/backuprepocleanupuc/backuprepocleanupdto"
)

const (
	getSettingRetryMax = 1
)

func (uc *UC) GetBackupRepoCleanup(
	ctx context.Context,
	auth *basedto.Auth,
	req *backuprepocleanupdto.GetBackupRepoCleanupReq,
) (_ *backuprepocleanupdto.GetBackupRepoCleanupResp, err error) {
	req.Type = currentSettingType
	var resp *settings.GetUniqueSettingResp
	for i := range getSettingRetryMax + 1 {
		resp, err = uc.GetUniqueSetting(ctx, auth, &req.GetUniqueSettingReq, &settings.GetUniqueSettingData{})
		if err == nil {
			break
		}
		if i < getSettingRetryMax && errors.Is(err, hperrors.ErrNotFound) {
			if e := uc.SettingInitService.InitDefaults(ctx, uc.DB); e != nil {
				return nil, hperrors.Wrap(err)
			}
			continue
		}
		return nil, hperrors.Wrap(err)
	}

	respData, err := backuprepocleanupdto.TransformBackupRepoCleanup(resp.Data, resp.RefObjects)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	return &backuprepocleanupdto.GetBackupRepoCleanupResp{
		Data: respData,
	}, nil
}
