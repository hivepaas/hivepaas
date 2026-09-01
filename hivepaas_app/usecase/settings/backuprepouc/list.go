package backuprepouc

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/backuprepouc/backuprepodto"
)

func (uc *UC) ListBackupRepo(
	ctx context.Context,
	auth *basedto.Auth,
	req *backuprepodto.ListBackupRepoReq,
) (*backuprepodto.ListBackupRepoResp, error) {
	req.Type = currentSettingType
	resp, err := uc.ListSetting(ctx, auth, &req.ListSettingReq, &settings.ListSettingData{})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	respData, err := backuprepodto.TransformBackupRepos(resp.Data, resp.RefObjects)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	return &backuprepodto.ListBackupRepoResp{
		Meta: resp.Meta,
		Data: respData,
	}, nil
}
