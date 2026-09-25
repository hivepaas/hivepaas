package settingmountuc

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/settingmountuc/settingmountdto"
)

// DeleteSettingMount deletes an entry. Its files leave the app with the refresh
// the delete event records; handing out less takes no gate.
func (uc *UC) DeleteSettingMount(
	ctx context.Context, auth *basedto.Auth, req *settingmountdto.DeleteSettingMountReq,
) (*settingmountdto.DeleteSettingMountResp, error) {
	req.Type = currentSettingType
	req.Auth = auth
	if _, err := uc.DeleteSetting(ctx, &req.DeleteSettingReq, &settings.DeleteSettingData{}); err != nil {
		return nil, hperrors.Wrap(err)
	}
	return &settingmountdto.DeleteSettingMountResp{}, nil
}
