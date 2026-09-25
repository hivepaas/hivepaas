package settings

import (
	"context"

	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/settingmountservice"
)

// CheckMountPaths refuses a file of an app at a path one of its other secrets,
// config files or setting mounts has. Only an app's settings are files of a
// container; an empty path is a setting with no file.
func (uc *BaseUC) CheckMountPaths(
	ctx context.Context, db database.IDB, scope *entity.ObjectScope, exceptSettingID string, paths ...string,
) error {
	paths = gofn.ToSliceSkippingZero(paths...)
	if scope == nil || !scope.IsAppScope() || len(paths) == 0 {
		return nil
	}
	claimed, err := uc.SettingMountService.ClaimedPaths(ctx, db, scope.AppID, exceptSettingID)
	if err != nil {
		return hperrors.Wrap(err)
	}
	return hperrors.Wrap(settingmountservice.CheckPathsFree(claimed, paths...))
}

// CheckMountPathsAfterLoading is CheckMountPaths as an update's AfterLoading
// hook: the setting being updated is loaded by then, and is left out.
func (uc *BaseUC) CheckMountPathsAfterLoading(
	scope *entity.ObjectScope, paths ...string,
) func(context.Context, database.Tx, *UpdateSettingData) error {
	return func(ctx context.Context, db database.Tx, data *UpdateSettingData) error {
		return uc.CheckMountPaths(ctx, db, scope, data.Setting.ID, paths...)
	}
}
