package periodicjobuc

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/taskservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
)

const (
	currentSettingType    = base.SettingTypePeriodicJob
	currentSettingVersion = entity.CurrentPeriodicJobVersion
)

type UC struct {
	taskService taskservice.Service

	*settings.BaseUC
}

func New(
	taskService taskservice.Service,

	baseUC *settings.BaseUC,
) *UC {
	return &UC{
		taskService: taskService,

		BaseUC: baseUC,
	}
}
