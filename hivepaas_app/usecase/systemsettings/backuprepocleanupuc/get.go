package backuprepocleanupuc

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/systemsettings/backuprepocleanupuc/backuprepocleanupdto"
)

func (uc *UC) GetBackupRepoCleanup(
	ctx context.Context,
	auth *basedto.Auth,
	req *backuprepocleanupdto.GetBackupRepoCleanupReq,
) (*backuprepocleanupdto.GetBackupRepoCleanupResp, error) {
	req.Type = currentSettingType
	resp, err := uc.GetUniqueSettingOrEmpty(ctx, auth, &req.GetUniqueSettingReq, &settings.GetUniqueSettingData{})
	if err != nil {
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
