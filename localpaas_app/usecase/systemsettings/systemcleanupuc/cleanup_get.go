package systemcleanupuc

import (
	"context"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings"
	"github.com/localpaas/localpaas/localpaas_app/usecase/systemsettings/systemcleanupuc/systemcleanupdto"
)

func (uc *UC) GetSystemCleanup(
	ctx context.Context,
	auth *basedto.Auth,
	req *systemcleanupdto.GetSystemCleanupReq,
) (*systemcleanupdto.GetSystemCleanupResp, error) {
	req.Type = currentSettingType
	resp, err := uc.GetSetting(ctx, auth, &req.GetSettingReq, &settings.GetSettingData{})
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	respData, err := systemcleanupdto.TransformSystemCleanup(resp.Data, resp.RefObjects)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	return &systemcleanupdto.GetSystemCleanupResp{
		Data: respData,
	}, nil
}
