package settingmountuc

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/settingmountservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/settingmountuc/settingmountdto"
)

func (uc *UC) GetSettingMount(
	ctx context.Context, auth *basedto.Auth, req *settingmountdto.GetSettingMountReq,
) (*settingmountdto.GetSettingMountResp, error) {
	req.Type = currentSettingType
	resp, err := uc.GetSetting(ctx, uc.DB, auth, &req.GetSettingReq, &settings.GetSettingData{})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	states := uc.entryStates(ctx, req.Scope)
	data, err := settingmountdto.TransformSettingMount(resp.Data, resp.RefObjects, states[resp.Data.ID])
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	return &settingmountdto.GetSettingMountResp{Data: data}, nil
}

// entryStates are what of each entry of the scope's app is mounted. Failing to
// read them is not failing the read: the entries come back with no state, which
// the screen shows as unknown.
func (uc *UC) entryStates(ctx context.Context, scope *entity.ObjectScope) map[string]*settingmountservice.EntryState {
	if scope == nil || scope.App == nil {
		return nil
	}
	states, err := uc.settingMountService.EntryStates(ctx, uc.DB, scope.App)
	if err != nil {
		return nil
	}
	return states
}
