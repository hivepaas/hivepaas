package settingmountuc

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/settingmountservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/settingmountuc/settingmountdto"
)

func (uc *UC) CreateSettingMount(
	ctx context.Context, auth *basedto.Auth, req *settingmountdto.CreateSettingMountReq,
) (*settingmountdto.CreateSettingMountResp, error) {
	req.Type = currentSettingType
	req.Auth = auth
	mount := req.ToEntity()
	resp, err := uc.CreateSetting(ctx, &req.CreateSettingReq, &settings.CreateSettingData{
		VerifyingName:   req.Name,
		VerifyingRefIDs: mount.GetRefObjectIDs(),
		Version:         currentSettingVersion,
		PrepareCreation: func(
			ctx context.Context, db database.Tx,
			_ *settings.CreateSettingData, pData *settings.PersistingSettingCreationData,
		) error {
			if err := uc.checkEntry(ctx, db, req.Scope, "", req.Name, mount); err != nil {
				return err
			}
			if err := pData.Setting.SetData(mount); err != nil {
				return hperrors.Wrap(err)
			}
			return uc.authorizeGrants(ctx, auth, req.Scope, nil, pData.Setting, base.AuditLogSourceAPICreate)
		},
	})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	return &settingmountdto.CreateSettingMountResp{Data: resp.Data}, nil
}

// checkEntry refuses an entry wrong for its source, or with a path another file
// of the app has. The source's existence and visibility from the app were
// checked with the references.
func (uc *UC) checkEntry(
	ctx context.Context, db database.IDB, scope *entity.ObjectScope,
	exceptSettingID, key string, mount *entity.AppSettingMount,
) error {
	if scope == nil || !scope.IsAppScope() {
		return hperrors.NewUnsupported("Setting mount outside an app")
	}
	source, err := uc.SettingRepo.GetByID(ctx, db, scope, "", mount.Source.ID, true)
	if err != nil {
		return hperrors.Wrap(err)
	}
	if err = settingmountservice.CheckEntry(key, mount, source.Type); err != nil {
		return hperrors.Wrap(err)
	}
	claimed, err := uc.settingMountService.ClaimedPaths(ctx, db, scope.AppID, exceptSettingID)
	if err != nil {
		return hperrors.Wrap(err)
	}
	paths := make([]string, 0, len(mount.Files))
	for _, f := range mount.Files {
		paths = append(paths, f.Path)
	}
	return hperrors.Wrap(settingmountservice.CheckPathsFree(claimed, paths...))
}
