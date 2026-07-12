package taskserviceimpl

import (
	"context"
	"fmt"
	"time"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/transaction"
)

func (s *service) LockAllPendingTasks(
	ctx context.Context,
	db database.Tx,
	maxWait time.Duration,
	extraOpts ...bunex.SelectQueryOption,
) ([]*entity.Task, error) {
	// Wait for at most when use SELECT FOR UPDATE
	if maxWait > 0 {
		_, err := db.Exec(fmt.Sprintf("SET LOCAL lock_timeout = '%vs';", int64(maxWait.Seconds())))
		if err != nil {
			return nil, apperrors.Wrap(err)
		}
	}

	// Lock all pending tasks from execution by the app and workers
	opts := []bunex.SelectQueryOption{
		bunex.SelectFor("UPDATE OF task"),
		bunex.SelectWhereIn("task.status IN (?)", base.TaskStatusNotStarted, base.TaskStatusInProgress),
		bunex.SelectColumns("id"),
	}
	opts = append(opts, extraOpts...)

	for {
		tasks, _, err := s.taskRepo.List(ctx, db, "", nil, opts...)
		if err == nil {
			return tasks, nil
		}
		if maxWait > 0 || !transaction.IsErrorDeadLock(err) {
			return nil, apperrors.Wrap(err)
		}
	}
}
