package settingmountuc

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/settingmountuc/settingmountdto"
)

func (uc *UC) ListSettingMount(
	ctx context.Context, auth *basedto.Auth, req *settingmountdto.ListSettingMountReq,
) (*settingmountdto.ListSettingMountResp, error) {
	req.Type = currentSettingType
	resp, err := uc.ListSetting(ctx, auth, &req.ListSettingReq, &settings.ListSettingData{})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	data, err := settingmountdto.TransformSettingMounts(resp.Data, resp.RefObjects, uc.entryStates(ctx, req.Scope))
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	return &settingmountdto.ListSettingMountResp{Meta: resp.Meta, Data: data}, nil
}
