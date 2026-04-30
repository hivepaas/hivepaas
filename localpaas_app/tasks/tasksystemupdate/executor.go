package tasksystemupdate

import (
	"context"
	"fmt"

	"github.com/tiendc/gofn"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/base"
	"github.com/localpaas/localpaas/localpaas_app/entity"
	"github.com/localpaas/localpaas/localpaas_app/infra/database"
	"github.com/localpaas/localpaas/localpaas_app/infra/logging"
	"github.com/localpaas/localpaas/localpaas_app/pkg/applog"
	"github.com/localpaas/localpaas/localpaas_app/pkg/bunex"
	"github.com/localpaas/localpaas/localpaas_app/pkg/transaction"
	"github.com/localpaas/localpaas/localpaas_app/repository"
	"github.com/localpaas/localpaas/localpaas_app/repository/cacherepository"
	"github.com/localpaas/localpaas/localpaas_app/service/lpappservice"
	"github.com/localpaas/localpaas/localpaas_app/service/notificationservice"
	"github.com/localpaas/localpaas/localpaas_app/service/settingservice"
	"github.com/localpaas/localpaas/localpaas_app/service/traefikservice"
	"github.com/localpaas/localpaas/localpaas_app/service/userservice"
	"github.com/localpaas/localpaas/localpaas_app/tasks/queue"
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
	userService         userservice.Service
	notificationService notificationservice.Service
}

func NewExecutor(
	taskQueue queue.TaskQueue,
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
		userService:         userService,
		notificationService: notificationService,
	}
	taskQueue.RegisterExecutor(base.TaskTypeSystemUpdate, e.execute)
	return e
}

type taskData struct {
	*queue.TaskExecData
	UpdateArgs            *entity.TaskSystemUpdateArgs
	UpdateOutput          *entity.TaskSystemUpdateOutput
	CurrentAppReplicas    *uint64
	CurrentWorkerReplicas *uint64

	LogStore     *applog.Store
	NotifMsgData *notificationservice.BaseMsgDataSystemUpdateNotification
}

func (e *Executor) execute(
	ctx context.Context,
	db database.Tx,
	task *queue.TaskExecData,
) (err error) {
	data := &taskData{
		TaskExecData: task,
		UpdateArgs:   gofn.Must(task.Task.ArgsAsSystemUpdate()),
		UpdateOutput: &entity.TaskSystemUpdateOutput{},
		LogStore:     applog.NewLocalStore(fmt.Sprintf("task:%s:log", task.Task.ID)),
	}
	data.SetOnPostTransaction(func() { e.onPostTransaction(data) }) //nolint

	defer func() {
		if err == nil {
			if r := recover(); r != nil {
				err = apperrors.NewPanic(fmt.Sprintf("%v", r))
			}
		}
		_ = e.saveLogs(ctx, db, data, true)
	}()

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

	err = e.updateSystemVersion(ctx, db, data)
	if err != nil {
		return apperrors.Wrap(err)
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

func (e *Executor) onPostTransaction(
	data *taskData,
) {
	ctx := context.Background()
	db := e.db

	// NOTE: We are now outside the transaction, need to reset some data before using them again
	data.LogStore = applog.NewLocalStore(data.LogStore.Key)

	defer func() {
		_ = e.saveLogs(ctx, db, data, false)
	}()

	if data.Task.IsDone() || data.Task.IsFailedCompletely() {
		err := e.notifyForSystemUpdate(ctx, db, data)
		if err != nil {
			_ = data.LogStore.Add(ctx, applog.NewOutFrame("Failed to send system update notification"+
				" with error: "+err.Error(), applog.TsNow))
		}
	}
}
