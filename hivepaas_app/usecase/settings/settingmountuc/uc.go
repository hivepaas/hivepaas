package settingmountuc

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/settingmountservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
)

const (
	currentSettingType    = base.SettingTypeAppSettingMount
	currentSettingVersion = entity.CurrentAppSettingMountVersion
)

// UC serves an app's setting mounts: parts of other settings - a certificate
// and its key, a basic auth pair as htpasswd - mounted as files. See
// docs/superpowers/specs/2026-09-25-setting-mounts-design.md.
type UC struct {
	*settings.BaseUC
	settingMountService settingmountservice.Service
}

func New(baseUC *settings.BaseUC, settingMountService settingmountservice.Service) *UC {
	return &UC{BaseUC: baseUC, settingMountService: settingMountService}
}
