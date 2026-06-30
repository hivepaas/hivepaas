package tasksystemupdate

import (
	"context"
	"errors"
	"fmt"

	"github.com/tiendc/gofn"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/base"
	"github.com/localpaas/localpaas/localpaas_app/entity"
	"github.com/localpaas/localpaas/localpaas_app/infra/database"
	"github.com/localpaas/localpaas/localpaas_app/infra/logging"
	"github.com/localpaas/localpaas/localpaas_app/pkg/bunex"
	"github.com/localpaas/localpaas/localpaas_app/pkg/funcutil"
	"github.com/localpaas/localpaas/localpaas_app/pkg/tasklog"
	"github.com/localpaas/localpaas/localpaas_app/pkg/timeutil"
	"github.com/localpaas/localpaas/localpaas_app/pkg/transaction"
	"github.com/localpaas/localpaas/localpaas_app/repository"
	"github.com/localpaas/localpaas/localpaas_app/service/sysupdateservice"
	"github.com/localpaas/localpaas/localpaas_app/service/taskservice"
	"github.com/localpaas/localpaas/localpaas_app/tasks/queue"
)

type Executor struct {
	logger    logging.Logger
	taskQueue queue.TaskQueue

	taskLogRepo repository.TaskLogRepo
	taskRepo    repository.TaskRepo

	sysUpdateService sysupdateservice.Service
	taskService      taskservice.Service
}

func NewExecutor(
	logger logging.Logger,
	taskQueue queue.TaskQueue,

	taskLogRepo repository.TaskLogRepo,
	taskRepo repository.TaskRepo,

	sysUpdateService sysupdateservice.Service,
	taskService taskservice.Service,
) *Executor {
	return &Executor{
		logger:    logger,
		taskQueue: taskQueue,

		taskLogRepo: taskLogRepo,
		taskRepo:    taskRepo,

		sysUpdateService: sysUpdateService,
		taskService:      taskService,
	}
}

type taskData struct {
	*queue.TaskExecData
}

func (e *Executor) Execute(
	ctx context.Context,
	db database.IDB,
) (err error) {
	task, err := e.loadUpdateTask(ctx, db)
	if err != nil {
		return apperrors.New(err)
	}
	if task == nil {
		return nil
	}

	data := &taskData{
		TaskExecData: &queue.TaskExecData{
			Task:       task,
			RefObjects: entity.NewRefObjects(),
			LogStore:   tasklog.NewLocalStore(fmt.Sprintf("task:%s:log", task.ID)),
		},
	}

	ctx, cancel := task.CreateTimeoutCtx(ctx)
	defer cancel()

	defer func() {
		if data.Task == nil {
			return
		}
		// Save task into the DB
		err2 := e.taskRepo.Update(ctx, db, data.Task)
		// Save all logs into the DB
		err3 := e.saveLogs(ctx, db, data, true)
		err = errors.Join(err, err2, err3)
	}()
	defer funcutil.EnsureNoPanic(&err) // Early catch panic before the above defers

	err = transaction.Execute(ctx, db, func(db database.Tx) error {
		// Lock all pending tasks from execution by the app and workers
		_, err = e.taskService.LockAllPendingTasks(ctx, db, 0)
		if err != nil {
			return apperrors.New(err)
		}

		latestTask, err := e.loadUpdateTask(ctx, db,
			bunex.SelectFor("UPDATE OF task"),
		)
		if err != nil {
			return apperrors.New(err)
		}
		data.Task = latestTask
		if latestTask == nil {
			return nil
		}

		// Stop only services which need to be stopped (main app and workers)
		err = e.taskQueue.StopAllSchedulers()
		if err != nil {
			return apperrors.New(err)
		}

		return nil
	})
	if err != nil {
		return apperrors.New(err)
	}
	if data.Task == nil {
		return nil
	}

	// NOTE: we do the system update outside the transaction as updating the DB
	// will restart the DB service, hence we can't keep it.

	_, err = e.sysUpdateService.SysUpdate(ctx, db, &sysupdateservice.SysUpdateReq{
		TaskExecData: data.TaskExecData,
	})
	if err != nil {
		return apperrors.New(err)
	}

	return nil
}

func (e *Executor) loadUpdateTask(
	ctx context.Context,
	db database.IDB,
	extraOpts ...bunex.SelectQueryOption,
) (*entity.Task, error) {
	opts := []bunex.SelectQueryOption{
		bunex.SelectWhere("task.type = ?", base.TaskTypeSystemUpdate),
		bunex.SelectWhere("task.status IN (?)", base.TaskStatusNotStarted),
		bunex.SelectWhere("(task.run_at IS NULL OR task.run_at < NOW())"),
		bunex.SelectOrder("created_at DESC"),
		bunex.SelectLimit(1),
	}
	opts = append(opts, extraOpts...)
	tasks, _, err := e.taskRepo.List(ctx, db, "", nil, opts...)
	if err != nil {
		return nil, apperrors.New(err)
	}
	if len(tasks) == 0 {
		return nil, nil
	}

	task := tasks[0]
	task.StartedAt = timeutil.NowUTC()
	if task.RunAt.IsZero() {
		task.RunAt = task.StartedAt
	}

	return task, nil
}

func (e *Executor) saveLogs(
	ctx context.Context,
	db database.IDB,
	data *taskData,
	addDurationInfo bool,
) error {
	task := data.Task
	logStore := data.LogStore
	if logStore == nil {
		return nil
	}

	if addDurationInfo {
		_ = logStore.Add(ctx, tasklog.NewOutFrame("System update finished in "+
			task.GetDuration().String(), tasklog.TsNow))
	}

	logFrames, err := logStore.GetData(ctx, 0)
	if err != nil {
		return apperrors.New(err)
	}
	_ = logStore.Close() //nolint

	// Insert data in to DB by chunk to avoid exceeding DBMS limit
	for _, chunk := range gofn.Chunk(logFrames, 10000) { //nolint
		taskLogs := make([]*entity.TaskLog, 0, len(chunk))
		for _, logFrame := range chunk {
			taskLogs = append(taskLogs, &entity.TaskLog{
				TaskID: data.Task.ID,
				Type:   logFrame.Type,
				Data:   logFrame.Data,
				Ts:     logFrame.Ts,
			})
		}
		err = e.taskLogRepo.InsertMulti(ctx, db, taskLogs)
		if err != nil {
			return apperrors.New(err)
		}
	}

	return nil
}
