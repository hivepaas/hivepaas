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

func (uc *UC) UpdateSettingMount(
	ctx context.Context, auth *basedto.Auth, req *settingmountdto.UpdateSettingMountReq,
) (*settingmountdto.UpdateSettingMountResp, error) {
	req.Type = currentSettingType
	req.Auth = auth
	mount := req.ToEntity()
	_, err := uc.UpdateSetting(ctx, &req.UpdateSettingReq, &settings.UpdateSettingData{
		VerifyingName:   req.Name,
		VerifyingRefIDs: mount.GetRefObjectIDs(),
		PrepareUpdate: func(
			ctx context.Context, db database.Tx,
			data *settings.UpdateSettingData, pData *settings.PersistingSettingData,
		) error {
			if err := uc.checkEntry(ctx, db, req.Scope, data.Setting.ID, req.Name, mount); err != nil {
				return err
			}
			if err := pData.Setting.SetData(mount); err != nil {
				return hperrors.Wrap(err)
			}
			return uc.authorizeGrants(ctx, auth, req.Scope, data.Setting, pData.Setting, base.AuditLogSourceAPIUpdate)
		},
	})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	return &settingmountdto.UpdateSettingMountResp{}, nil
}
