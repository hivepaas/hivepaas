package backuprepouc

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/backuprepouc/backuprepodto"
)

func (uc *UC) GetBackupRepo(
	ctx context.Context,
	auth *basedto.Auth,
	req *backuprepodto.GetBackupRepoReq,
) (*backuprepodto.GetBackupRepoResp, error) {
	req.Type = currentSettingType
	resp, err := uc.GetSetting(ctx, auth, &req.GetSettingReq, &settings.GetSettingData{})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	setting := resp.Data
	if setting.ObjectID == setting.CurrentObjectID { // not return sensitive data if setting is inherited
		if err := setting.MustAsBackupRepo().Decrypt(); err != nil {
			return nil, hperrors.Wrap(err)
		}
	}

	respData, err := backuprepodto.TransformBackupRepo(setting, resp.RefObjects)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	return &backuprepodto.GetBackupRepoResp{
		Data: respData,
	}, nil
}
