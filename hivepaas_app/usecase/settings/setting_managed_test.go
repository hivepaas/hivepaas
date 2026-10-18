package settings

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
)

func managedSettings() []*entity.Setting {
	managed := &entity.Setting{Type: base.SettingTypeAppManaged, Status: base.SettingStatusActive}
	managed.MustSetData(&entity.AppManagedSettings{Engine: "postgres",
		Locks: entity.ManagedLocks{Settings: []string{"secret-1", "script-1"}}})
	return []*entity.Setting{managed}
}

// The credentials and scripts a managed app was built with are its engine's:
// editing, pausing or deleting one is refused, naming the type.
func TestAManagedAppsOwnSettingsAreLocked(t *testing.T) {
	for id, typ := range map[string]base.SettingType{
		"secret-1": base.SettingTypeSecret, "script-1": base.SettingTypeConfigFile,
	} {
		setting := &entity.Setting{ID: id, Type: typ, Scope: base.ObjectScopeApp, ObjectID: "app-1"}

		err := managedSettingLocked(setting, managedSettings())

		assert.ErrorIs(t, err, hperrors.ErrAppManagedFieldLocked, id)
		info, _ := hperrors.ParseError(err, "en")
		assert.Equal(t, string(typ), info.Params["Field"], id)
	}
}

func TestOtherSettingsAreNotLocked(t *testing.T) {
	for name, setting := range map[string]*entity.Setting{
		"another secret of the app":     {ID: "secret-2", Type: base.SettingTypeSecret, Scope: base.ObjectScopeApp},
		"a project's secret":            {ID: "secret-1", Type: base.SettingTypeSecret, Scope: base.ObjectScopeProject},
		"a type the engine never locks": {ID: "secret-1", Type: base.SettingTypeEnvVar, Scope: base.ObjectScopeApp},
	} {
		assert.NoError(t, managedSettingLocked(setting, managedSettings()), name)
	}
	assert.NoError(t, managedSettingLocked(
		&entity.Setting{ID: "secret-1", Type: base.SettingTypeSecret, Scope: base.ObjectScopeApp}, nil),
		"an app no engine looks after")
}
