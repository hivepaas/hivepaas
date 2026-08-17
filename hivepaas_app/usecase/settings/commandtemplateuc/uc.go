package commandtemplateuc

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/commandservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
)

const (
	currentSettingType    = base.SettingTypeCommandTemplate
	currentSettingVersion = entity.CurrentCommandTemplateVersion
)

type UC struct {
	*settings.BaseUC
	commandService commandservice.Service
}

func New(
	baseUC *settings.BaseUC,
	commandService commandservice.Service,
) *UC {
	return &UC{
		BaseUC:         baseUC,
		commandService: commandService,
	}
}
