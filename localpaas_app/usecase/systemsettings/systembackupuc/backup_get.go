package systembackupuc

import (
	"context"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings"
	"github.com/localpaas/localpaas/localpaas_app/usecase/systemsettings/systembackupuc/systembackupdto"
)

func (uc *UC) GetSystemBackup(
	ctx context.Context,
	auth *basedto.Auth,
	req *systembackupdto.GetSystemBackupReq,
) (*systembackupdto.GetSystemBackupResp, error) {
	req.Type = currentSettingType
	resp, err := uc.GetSetting(ctx, auth, &req.GetSettingReq, &settings.GetSettingData{})
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	respData, err := systembackupdto.TransformSystemBackup(resp.Data, resp.RefObjects)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	return &systembackupdto.GetSystemBackupResp{
		Data: respData,
	}, nil
}
