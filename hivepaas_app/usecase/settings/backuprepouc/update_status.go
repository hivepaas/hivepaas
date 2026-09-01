package backuprepouc

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/backuprepouc/backuprepodto"
)

func (uc *UC) UpdateBackupRepoStatus(
	ctx context.Context,
	auth *basedto.Auth,
	req *backuprepodto.UpdateBackupRepoStatusReq,
) (*backuprepodto.UpdateBackupRepoStatusResp, error) {
	req.Type = currentSettingType
	_, err := uc.UpdateSettingStatus(ctx, &req.UpdateSettingStatusReq, &settings.UpdateSettingStatusData{})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	return &backuprepodto.UpdateBackupRepoStatusResp{}, nil
}
