package healthcheckuc

import (
	"github.com/localpaas/localpaas/localpaas_app/base"
	"github.com/localpaas/localpaas/localpaas_app/entity"
	"github.com/localpaas/localpaas/localpaas_app/service/taskservice"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings"
)

const (
	currentSettingType    = base.SettingTypeHealthcheck
	currentSettingVersion = entity.CurrentHealthcheckVersion
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
