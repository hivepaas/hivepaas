package settinghelper

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
)

func FindSettingsByScope(
	settings []*entity.Setting,
	scope *entity.ObjectScope,
	typ base.SettingType,
	kind *string,
) (res []*entity.Setting) {
	var settingsInParentApp, settingsInProjectEnv, settingsInProject, settingsInGlobal []*entity.Setting
	scopeObjectID := scope.ScopeObjectID()
	for _, setting := range settings {
		if setting.Type != typ || (kind != nil && setting.Kind != *kind) {
			continue
		}
		if setting.ObjectID == "" {
			settingsInGlobal = append(settingsInGlobal, setting)
			continue
		}
		if setting.ObjectID == scopeObjectID {
			res = append(res, setting) // found a setting in current scope
			continue
		}
		switch scope.ScopeType {
		case base.ObjectScopeApp:
			if setting.ObjectID == scope.ParentAppID {
				settingsInParentApp = append(settingsInParentApp, setting)
			}
			if setting.ObjectID == scope.ProjectEnvID {
				settingsInProjectEnv = append(settingsInProjectEnv, setting)
			}
			if setting.ObjectID == scope.ProjectID {
				settingsInProject = append(settingsInProject, setting)
			}
			continue
		case base.ObjectScopeProjectEnv:
			if setting.ObjectID == scope.ProjectID {
				settingsInProject = append(settingsInProject, setting)
			}
			continue
		case base.ObjectScopeProject:
		case base.ObjectScopeUser:
		case base.ObjectScopeGlobal:
		}
	}
	res = append(res, settingsInParentApp...)
	res = append(res, settingsInProjectEnv...)
	res = append(res, settingsInProject...)
	res = append(res, settingsInGlobal...)
	return res
}
