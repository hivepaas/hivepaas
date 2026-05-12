package localpaassettingsuc

import (
	"github.com/localpaas/localpaas/localpaas_app/base"
	"github.com/localpaas/localpaas/localpaas_app/service/lpappservice"
	"github.com/localpaas/localpaas/localpaas_app/service/taskservice"
	"github.com/localpaas/localpaas/localpaas_app/tasks/queue"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings"
	"github.com/localpaas/localpaas/services/docker"
)

const (
	currentSettingType = base.SettingTypeLocalPaaSSettings
)

type UC struct {
	*settings.BaseUC
	lpAppService  lpappservice.Service
	taskService   taskservice.Service
	taskQueue     queue.TaskQueue
	dockerManager docker.Manager
}

func New(
	baseUC *settings.BaseUC,
	lpAppService lpappservice.Service,
	taskService taskservice.Service,
	taskQueue queue.TaskQueue,
	dockerManager docker.Manager,
) *UC {
	return &UC{
		BaseUC:        baseUC,
		lpAppService:  lpAppService,
		taskService:   taskService,
		taskQueue:     taskQueue,
		dockerManager: dockerManager,
	}
}
