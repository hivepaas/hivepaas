package taskservice

import (
	"context"
	"time"

	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
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

	// CancelTask stops a task the caller is entitled to stop.
	//
	// scope is what the caller reached the task through - the project, the
	// environment, the app - and the task has to be inside it, inheritance
	// included. A nil scope means no such check, which is only for callers that
	// have already proven the task is theirs by other means.
	//
	// The task it returns is the row as it stood before the cancel, so a caller
	// can record what was stopped.
	CancelTask(ctx context.Context, db database.Tx, scope *entity.ObjectScope, taskID string,
		validatingTargetID *string) (task *entity.Task, canceled bool, _ error)
	CancelInProgressTask(ctx context.Context, taskID string) error
}
