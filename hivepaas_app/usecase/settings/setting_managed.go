package settings

import (
	"context"
	"slices"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
)

// refuseManagedSetting refuses to edit, pause or delete a secret, config file or
// setting mount a managed app was built with: its engine's credentials and
// scripts, which its locks name. Every other setting is passed over without a
// read.
func (uc *BaseUC) refuseManagedSetting(ctx context.Context, db database.IDB, setting *entity.Setting) error {
	if !lockableByEngine(setting) {
		return nil
	}
	managed, _, err := uc.SettingRepo.List(ctx, db, nil, nil,
		bunex.SelectWhere("setting.type = ?", base.SettingTypeAppManaged),
		bunex.SelectWhere("setting.status = ?", base.SettingStatusActive),
		bunex.SelectWhere("setting.object_id = ?", setting.ObjectID),
	)
	if err != nil {
		return hperrors.Wrap(err)
	}
	return managedSettingLocked(setting, managed)
}

func lockableByEngine(setting *entity.Setting) bool {
	if setting == nil || setting.Scope != base.ObjectScopeApp {
		return false
	}
	switch setting.Type { //nolint:exhaustive // the three types an engine builds an app with
	case base.SettingTypeSecret, base.SettingTypeConfigFile, base.SettingTypeAppSettingMount:
		return true
	}
	return false
}

// managedSettingLocked is the refusal of a setting the app's engine owns, given
// the app's app-managed setting, if any.
func managedSettingLocked(setting *entity.Setting, managed []*entity.Setting) error {
	if !lockableByEngine(setting) {
		return nil
	}
	for _, one := range managed {
		if one.Type != base.SettingTypeAppManaged {
			continue
		}
		data, err := one.AsAppManagedSettings()
		if err != nil {
			return hperrors.Wrap(err)
		}
		if slices.Contains(data.Locks.Settings, setting.ID) {
			return hperrors.Wrap(hperrors.ErrAppManagedFieldLocked).WithParam("Field", string(setting.Type)).
				ExposeParams("Field")
		}
	}
	return nil
}
