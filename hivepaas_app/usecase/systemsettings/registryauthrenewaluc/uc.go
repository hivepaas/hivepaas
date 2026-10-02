package registryauthrenewaluc

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/repository"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/registryauthservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/tasks/queue"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
)

type UC struct {
	taskQueue queue.TaskQueue

	taskRepo repository.TaskRepo

	registryAuthService registryauthservice.Service

	*settings.BaseUC
}

func New(
	taskQueue queue.TaskQueue,

	taskRepo repository.TaskRepo,

	registryAuthService registryauthservice.Service,

	baseUC *settings.BaseUC,
) *UC {
	return &UC{
		taskQueue: taskQueue,

		taskRepo: taskRepo,

		registryAuthService: registryAuthService,

		BaseUC: baseUC,
	}
}
