package systemcleanupuc

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/systemsettings/systemcleanupuc/systemcleanupdto"
)

func (uc *UC) GetSystemCleanup(
	ctx context.Context,
	auth *basedto.Auth,
	req *systemcleanupdto.GetSystemCleanupReq,
) (*systemcleanupdto.GetSystemCleanupResp, error) {
	req.Type = currentSettingType
	resp, err := uc.GetUniqueSetting(ctx, auth, &req.GetUniqueSettingReq, &settings.GetUniqueSettingData{})
	if err != nil {
		return nil, apperrors.New(err)
	}

	respData, err := systemcleanupdto.TransformSystemCleanup(resp.Data, resp.RefObjects)
	if err != nil {
		return nil, apperrors.New(err)
	}

	return &systemcleanupdto.GetSystemCleanupResp{
		Data: respData,
	}, nil
}
