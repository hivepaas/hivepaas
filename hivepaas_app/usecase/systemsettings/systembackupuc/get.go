package systembackupuc

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/systemsettings/systembackupuc/systembackupdto"
)

func (uc *UC) GetSystemBackup(
	ctx context.Context,
	auth *basedto.Auth,
	req *systembackupdto.GetSystemBackupReq,
) (*systembackupdto.GetSystemBackupResp, error) {
	req.Type = currentSettingType
	resp, err := uc.GetUniqueSettingOrEmpty(ctx, auth, &req.GetUniqueSettingReq, &settings.GetUniqueSettingData{})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	respData, err := systembackupdto.TransformSystemBackup(resp.Data, resp.RefObjects)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	return &systembackupdto.GetSystemBackupResp{
		Data: respData,
	}, nil
}
