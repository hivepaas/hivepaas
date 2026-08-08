package entityutil

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
)

func GetSettingsByType(settings []*entity.Setting, typ base.SettingType) (resp []*entity.Setting) {
	for _, setting := range settings {
		if setting.Type == typ {
			resp = append(resp, setting)
		}
	}
	return resp
}

func GetSettingByType(settings []*entity.Setting, typ base.SettingType) *entity.Setting {
	for _, setting := range settings {
		if setting.Type == typ {
			return setting
		}
	}
	return nil
}
