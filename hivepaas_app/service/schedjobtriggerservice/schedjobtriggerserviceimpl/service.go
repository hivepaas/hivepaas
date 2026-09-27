package schedjobtriggerserviceimpl

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/transaction"
	"github.com/hivepaas/hivepaas/hivepaas_app/repository"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/schedjobservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/schedjobtriggerservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/taskservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/tasks/queue"
)

type service struct {
	db *database.DB

	settingRepo repository.SettingRepo
	taskRepo    repository.TaskRepo

	schedJobService schedjobservice.Service
	taskQueue       queue.TaskQueue

	// cancelRun cancels a run a canceled deploy waited for.
	cancelRun func(ctx context.Context, taskID string) error
}

func New(
	db *database.DB,
	settingRepo repository.SettingRepo,
	taskRepo repository.TaskRepo,
	schedJobService schedjobservice.Service,
	taskService taskservice.Service,
	taskQueue queue.TaskQueue,
) schedjobtriggerservice.Service {
	svc := &service{
		db:              db,
		settingRepo:     settingRepo,
		taskRepo:        taskRepo,
		schedJobService: schedJobService,
		taskQueue:       taskQueue,
	}
	svc.cancelRun = func(ctx context.Context, taskID string) error {
		return transaction.Execute(ctx, db, func(tx database.Tx) error {
			_, _, err := taskService.CancelTask(ctx, tx, nil, taskID, nil)
			return hperrors.Wrap(err)
		})
	}
	return svc
}
