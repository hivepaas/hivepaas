package taskservice

import (
	"context"
	"time"

	"github.com/localpaas/localpaas/localpaas_app/entity"
	"github.com/localpaas/localpaas/localpaas_app/infra/database"
	"github.com/localpaas/localpaas/localpaas_app/pkg/bunex"
)

type Service interface {
	GetTask(ctx context.Context, db database.IDB, req *GetTaskReq) (*GetTaskResp, error)
	ListTask(ctx context.Context, db database.IDB, req *ListTaskReq) (*ListTaskResp, error)

	// Logs
	GetTaskLogs(ctx context.Context, db database.IDB, req *GetTaskLogsReq) (*GetTaskLogsResp, error)

	// Locking
	CreateDBLock(ctx context.Context, db database.Tx, id, selectFor string) (*entity.Lock, error)
	CreateRedisLock(ctx context.Context, key string, exp time.Duration) (success bool, releaser func(), err error)
	LockAllPendingTasks(ctx context.Context, db database.Tx, maxWait time.Duration,
		extraOpts ...bunex.SelectQueryOption) ([]*entity.Task, error)

	// Cancel a task
	CancelTask(ctx context.Context, db database.Tx, taskID string, validatingTargetID *string) (canceled bool, _ error)
	CancelInProgressTask(ctx context.Context, taskID string) error
}
