package systembackupuc

import (
	"context"
	"errors"

	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/systemsettings/systembackupuc/systembackupdto"
)

const (
	getSettingRetryMax = 1
)

func (uc *UC) GetSystemBackup(
	ctx context.Context,
	auth *basedto.Auth,
	req *systembackupdto.GetSystemBackupReq,
) (_ *systembackupdto.GetSystemBackupResp, err error) {
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

	respData, err := systembackupdto.TransformSystemBackup(resp.Data, resp.RefObjects)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	return &systembackupdto.GetSystemBackupResp{
		Data: respData,
	}, nil
}
