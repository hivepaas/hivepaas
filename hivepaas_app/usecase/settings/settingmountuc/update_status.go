package settingmountuc

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/settingmountuc/settingmountdto"
)

// UpdateSettingMountStatus enables or disables an entry. Enabling one hands its
// files out again: its paths are checked afresh, since another file may have
// taken one meanwhile, and a sensitive part passes the gate.
func (uc *UC) UpdateSettingMountStatus(
	ctx context.Context, auth *basedto.Auth, req *settingmountdto.UpdateSettingMountStatusReq,
) (*settingmountdto.UpdateSettingMountStatusResp, error) {
	req.Type = currentSettingType
	req.Auth = auth
	_, err := uc.UpdateSettingStatus(ctx, &req.UpdateSettingStatusReq, &settings.UpdateSettingStatusData{
		BeforePersisting: func(
			ctx context.Context, db database.Tx,
			data *settings.UpdateSettingStatusData, pData *settings.PersistingSettingStatusData,
		) error {
			if !pData.Setting.IsActive() {
				return nil
			}
			mount, err := pData.Setting.AsAppSettingMount()
			if err != nil {
				return hperrors.Wrap(err)
			}
			if err = uc.checkEntry(ctx, db, req.Scope, pData.Setting.ID, pData.Setting.Name, mount); err != nil {
				return err
			}
			return uc.authorizeGrants(ctx, auth, req.Scope, data.Setting, pData.Setting, base.AuditLogSourceAPIUpdate)
		},
	})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	return &settingmountdto.UpdateSettingMountStatusResp{}, nil
}
