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
	"github.com/localpaas/localpaas/localpaas_app/pkg/applog"
	"github.com/localpaas/localpaas/localpaas_app/pkg/bunex"
	"github.com/localpaas/localpaas/localpaas_app/pkg/funcutil"
	"github.com/localpaas/localpaas/localpaas_app/pkg/timeutil"
	"github.com/localpaas/localpaas/localpaas_app/pkg/transaction"
	"github.com/localpaas/localpaas/localpaas_app/repository"
	"github.com/localpaas/localpaas/localpaas_app/repository/cacherepository"
	"github.com/localpaas/localpaas/localpaas_app/service/dbservice"
	"github.com/localpaas/localpaas/localpaas_app/service/lpappservice"
	"github.com/localpaas/localpaas/localpaas_app/service/notificationservice"
	"github.com/localpaas/localpaas/localpaas_app/service/settingservice"
	"github.com/localpaas/localpaas/localpaas_app/service/traefikservice"
	"github.com/localpaas/localpaas/localpaas_app/service/userservice"
	"github.com/localpaas/localpaas/services/docker"
)

type Executor struct {
	logger              logging.Logger
	db                  *database.DB
	settingRepo         repository.SettingRepo
	taskLogRepo         repository.TaskLogRepo
	taskRepo            repository.TaskRepo
	taskInfoRepo        cacherepository.TaskInfoRepo
	dockerManager       docker.Manager
	settingService      settingservice.Service
	lpAppService        lpappservice.Service
	traefikService      traefikservice.Service
	dbService           dbservice.Service
	userService         userservice.Service
	notificationService notificationservice.Service
}

func NewExecutor(
	logger logging.Logger,
	db *database.DB,
	settingRepo repository.SettingRepo,
	taskLogRepo repository.TaskLogRepo,
	taskRepo repository.TaskRepo,
	taskInfoRepo cacherepository.TaskInfoRepo,
	dockerManager docker.Manager,
	settingService settingservice.Service,
	lpAppService lpappservice.Service,
	traefikService traefikservice.Service,
	dbService dbservice.Service,
	userService userservice.Service,
	notificationService notificationservice.Service,
) *Executor {
	e := &Executor{
		logger:              logger,
		db:                  db,
		settingRepo:         settingRepo,
		taskLogRepo:         taskLogRepo,
		taskRepo:            taskRepo,
		taskInfoRepo:        taskInfoRepo,
		dockerManager:       dockerManager,
		settingService:      settingService,
		lpAppService:        lpAppService,
		traefikService:      traefikService,
		dbService:           dbService,
		userService:         userService,
		notificationService: notificationService,
	}
	return e
}

type taskData struct {
	Task       *entity.Task
	RefObjects *entity.RefObjects

	UpdateArgs            *entity.TaskSystemUpdateArgs
	UpdateOutput          *entity.TaskSystemUpdateOutput
	CurrentAppReplicas    *uint64
	CurrentWorkerReplicas *uint64

	LogStore     *applog.Store
	NotifMsgData *notificationservice.TemplateDataSystemUpdate
}

func (e *Executor) Execute(
	ctx context.Context,
	db database.IDB,
) (err error) {
	defer funcutil.EnsureNoPanic(&err) // Make sure no panic at all

	task, err := e.loadUpdateTask(ctx, db)
	if err != nil {
		return apperrors.Wrap(err)
	}
	if task == nil {
		return nil
	}

	data := &taskData{
		Task:         task,
		RefObjects:   &entity.RefObjects{},
		UpdateArgs:   gofn.Must(task.ArgsAsSystemUpdate()),
		UpdateOutput: &entity.TaskSystemUpdateOutput{},
		LogStore:     applog.NewLocalStore(fmt.Sprintf("task:%s:log", task.ID)),
	}

	ctx, cancel := task.CreateTimeoutCtx(ctx)
	defer cancel()

	defer func() {
		if data.Task == nil {
			return
		}
		// Save update task
		err = e.onAfterSystemUpdate(ctx, data)
	}()
	defer func() {
		if data.Task == nil {
			return
		}
		// Save update task
		err1 := e.saveUpdateTask(ctx, db, err, data)
		// Send result notifications
		err2 := e.sendResultNotifications(ctx, db, data)
		// Save all logs of the processing
		err3 := e.saveLogs(ctx, db, data, true)
		// CRITICAL: Must include the existing 'err' so we don't drop panic errors or update errors!
		err = errors.Join(err, err1, err2, err3)
	}()
	defer funcutil.EnsureNoPanic(&err) // Early catch panic before the above defers

	err = transaction.Execute(ctx, db, func(db database.Tx) error {
		// Lock all pending tasks from execution by the app and workers
		err = e.lockAllPendingTasks(ctx, db)
		if err != nil {
			return apperrors.Wrap(err)
		}

		latestTask, err := e.loadUpdateTask(ctx, db,
			bunex.SelectFor("UPDATE OF task"),
		)
		if err != nil {
			return apperrors.Wrap(err)
		}
		data.Task = latestTask
		if latestTask == nil {
			return nil
		}

		// Stop only services which need to be stopped (main app and workers)
		err = e.stopServices(ctx, data)
		if err != nil {
			return apperrors.Wrap(err)
		}

		return nil
	})
	if err != nil {
		return apperrors.Wrap(err)
	}

	// NOTE: we do the system update outside the transaction as updating the DB
	// will restart the DB service, hence we can't keep it.

	err = e.onBeforeSystemUpdate(ctx, data)
	if err != nil {
		return apperrors.Wrap(err)
	}

	err = e.updateSystem(ctx, db, data)
	if err != nil {
		return apperrors.Wrap(err)
	}

	return nil
}

func (e *Executor) lockAllPendingTasks(
	ctx context.Context,
	db database.Tx,
) error {
	// Lock all pending tasks from execution by the app and workers
	for {
		_, _, err := e.taskRepo.List(ctx, db, "", nil,
			bunex.SelectFor("UPDATE OF task"),
			bunex.SelectWhereIn("task.status IN (?)", base.TaskStatusNotStarted, base.TaskStatusInProgress),
			bunex.SelectColumns("id"),
		)
		if err == nil {
			break
		}
		if !transaction.IsErrorDeadLock(err) {
			return apperrors.Wrap(err)
		}
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
		return nil, apperrors.Wrap(err)
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

func (e *Executor) saveUpdateTask(
	ctx context.Context,
	db database.IDB,
	updateErr error,
	data *taskData,
) error {
	task := data.Task
	task.EndedAt = timeutil.NowUTC()
	if updateErr != nil {
		task.Status = base.TaskStatusFailed
		_ = task.AddRun(&entity.TaskRun{
			StartedAt: task.StartedAt,
			EndedAt:   task.EndedAt,
			Error:     updateErr.Error(),
		})
	} else {
		task.Status = base.TaskStatusDone
	}
	// Save task in DB
	err := e.taskRepo.Update(ctx, db, task)
	if err != nil {
		return apperrors.Wrap(err)
	}
	return nil
}

func (e *Executor) sendResultNotifications(
	ctx context.Context,
	db database.IDB,
	data *taskData,
) error {
	task := data.Task
	if task.IsDone() || task.IsFailedCompletely() {
		err := e.notifyForSystemUpdate(ctx, db, data)
		if err != nil {
			_ = data.LogStore.Add(ctx, applog.NewOutFrame("Failed to send system update notification"+
				" with error: "+err.Error(), applog.TsNow))
			return apperrors.Wrap(err)
		}
	}
	return nil
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
		_ = logStore.Add(ctx, applog.NewOutFrame("System update finished in "+
			task.GetDuration().String(), applog.TsNow))
	}

	logFrames, err := logStore.GetData(ctx, 0)
	if err != nil {
		return apperrors.Wrap(err)
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
			return apperrors.Wrap(err)
		}
	}

	return nil
}
