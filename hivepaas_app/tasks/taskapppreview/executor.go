package taskapppreview

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/apppreviewservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/tasks/queue"
)

type Executor struct {
	appPreviewService apppreviewservice.Service
}

func NewExecutor(
	taskQueue queue.TaskQueue,
	appPreviewService apppreviewservice.Service,
) *Executor {
	e := &Executor{
		appPreviewService: appPreviewService,
	}
	taskQueue.RegisterExecutor(base.TaskTypeAppPreview, e.execute)
	return e
}

func (e *Executor) execute(
	ctx context.Context,
	db database.Tx,
	execData *queue.TaskExecData,
) error {
	_, err := e.appPreviewService.CreatePreview(ctx, db, &apppreviewservice.CreatePreviewReq{
		TaskExecData: execData,
	})
	if err != nil {
		return hperrors.Wrap(err)
	}
	return nil
}
