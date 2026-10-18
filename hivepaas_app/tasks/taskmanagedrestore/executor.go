// Package taskmanagedrestore sees a managed database restored into a new app
// through, after the app is made (design §4, "The steps"): it waits for the
// recovery to end, has the app take its first base backup, and takes off it
// what it was restored with.
//
// Each run of the task does one step, or waits a while on one, and continues
// the task: the step it is at is in the task's output, so a run that crashed
// is resumed where it was, and each step can run again.
package taskmanagedrestore

import (
	"context"
	"fmt"
	"time"

	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/rediscache"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/safego"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/tasklog"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/repository"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/appservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/apptemplateservice/templatemodel"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/envvarservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/managedservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/managedservice/managedmodel"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/schedjobexecservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/schedjobservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/settingmountservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/tasks/queue"
	"github.com/hivepaas/hivepaas/services/docker"
)

type Executor struct {
	db          *database.DB
	redisClient rediscache.Client
	taskQueue   queue.TaskQueue

	settingRepo repository.SettingRepo
	taskLogRepo repository.TaskLogRepo
	taskRepo    repository.TaskRepo

	appService          appservice.Service
	envVarService       envvarservice.Service
	managedService      managedservice.Service
	schedJobExecService schedjobexecservice.Service
	schedJobService     schedjobservice.Service
	settingMountService settingmountservice.Service
	dockerManager       docker.Manager

	now   func() time.Time
	sleep func(ctx context.Context, d time.Duration) error
}

func NewExecutor(
	db *database.DB,
	redisClient rediscache.Client,
	taskQueue queue.TaskQueue,
	settingRepo repository.SettingRepo,
	taskLogRepo repository.TaskLogRepo,
	taskRepo repository.TaskRepo,
	appService appservice.Service,
	envVarService envvarservice.Service,
	managedService managedservice.Service,
	schedJobExecService schedjobexecservice.Service,
	schedJobService schedjobservice.Service,
	settingMountService settingmountservice.Service,
	dockerManager docker.Manager,
) *Executor {
	e := &Executor{
		db:                  db,
		redisClient:         redisClient,
		taskQueue:           taskQueue,
		settingRepo:         settingRepo,
		taskLogRepo:         taskLogRepo,
		taskRepo:            taskRepo,
		appService:          appService,
		envVarService:       envVarService,
		managedService:      managedService,
		schedJobExecService: schedJobExecService,
		schedJobService:     schedJobService,
		settingMountService: settingMountService,
		dockerManager:       dockerManager,
		now:                 timeutil.NowUTC,
		sleep:               sleepCtx,
	}
	taskQueue.RegisterExecutor(base.TaskTypeManagedRestore, e.execute)
	return e
}

func sleepCtx(ctx context.Context, d time.Duration) error {
	timer := time.NewTimer(d)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return hperrors.Wrap(ctx.Err())
	case <-timer.C:
		return nil
	}
}

// restoreRun is the restore a run of the task does a step of: the app restored
// into, its engine's setting, the engine, and where the restore is.
type restoreRun struct {
	data    *queue.TaskExecData
	app     *entity.App
	setting *entity.Setting
	managed *entity.AppManagedSettings
	engine  managedmodel.Engine
	state   *entity.TaskManagedRestoreState
}

func (e *Executor) execute(ctx context.Context, db database.Tx, data *queue.TaskExecData) (err error) {
	data.LogStore = tasklog.NewRemoteStore(fmt.Sprintf("task:%s:log", data.Task.ID), e.redisClient)
	data.LogStore.SetOnFlush(tasklog.DefaultMaxSize, func(ctx context.Context, frames []*tasklog.LogFrame) error {
		return e.saveLogFrames(ctx, e.db, data.Task, frames)
	})
	data.OnPostTx(func() { _ = e.saveLogs(context.Background(), e.db, data) }) //nolint:contextcheck
	defer func() {
		if err != nil {
			_ = data.LogStore.Add(ctx, tasklog.NewErrFrame("Restore failed: "+err.Error()+"\n", tasklog.TsNow))
		}
		_ = e.saveLogs(ctx, db, data)
	}()
	defer safego.RecoverTo(&err)

	run, err := e.loadRun(ctx, db, data)
	if err != nil {
		return err
	}
	err = e.runStep(ctx, db, run)
	data.Task.MustSetOutput(run.state)
	return err
}

func (e *Executor) runStep(ctx context.Context, db database.Tx, run *restoreRun) error {
	switch run.state.Step {
	case entity.ManagedRestoreStepRecover:
		return e.waitRecovery(ctx, db, run)
	case entity.ManagedRestoreStepBackup:
		return e.startBackup(ctx, db, run)
	case entity.ManagedRestoreStepWaitBackup:
		return e.waitBackup(ctx, db, run)
	case entity.ManagedRestoreStepFinish:
		return e.finish(ctx, db, run)
	case entity.ManagedRestoreStepDone:
		return nil
	}
	run.data.TaskNonRetryable = true
	return hperrors.Wrap(hperrors.ErrInternal).WithMsgLog("managed restore: unknown step %q", run.state.Step)
}

// loadRun loads what the restore is of. An app gone - deleted by somebody while
// it restored - ends the restore: there is nothing left to see through.
func (e *Executor) loadRun(ctx context.Context, db database.Tx, data *queue.TaskExecData) (*restoreRun, error) {
	args, err := data.Task.ArgsAsManagedRestore()
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	if args == nil {
		data.TaskNonRetryable = true
		return nil, hperrors.Wrap(hperrors.ErrInternal).WithMsgLog("managed restore task has no arguments")
	}
	state, err := data.Task.OutputAsManagedRestore()
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	if state == nil || state.Step == "" {
		state = &entity.TaskManagedRestoreState{Step: entity.ManagedRestoreStepRecover}
	}
	app, err := e.appService.LoadApp(ctx, db, args.ProjectID, args.AppID, false, false,
		bunex.SelectRelation("Project", bunex.SelectExcludeColumns(entity.ProjectDefaultExcludeColumns...)),
		bunex.SelectRelation("ProjectEnv"),
	)
	if err != nil {
		data.TaskNonRetryable = true
		return nil, hperrors.Wrap(err).WithExtraDetail("the app restored into is gone")
	}
	setting, err := e.settingRepo.GetSingle(ctx, db, app.GetObjectScope(), base.SettingTypeAppManaged, true,
		bunex.SelectWhere("setting.object_id = ?", app.ID), bunex.SelectFor("UPDATE OF setting"))
	if err != nil {
		data.TaskNonRetryable = true
		return nil, hperrors.Wrap(err).WithExtraDetail("the app restored into is not managed")
	}
	managed, err := setting.AsAppManagedSettings()
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	engine, found := e.managedService.Engine(templatemodel.ManagedEngine(managed.Engine))
	if !found {
		data.TaskNonRetryable = true
		return nil, hperrors.Wrap(hperrors.ErrAppTemplateIncompatible).
			WithExtraDetail("engine %s is not one this HivePaaS has", managed.Engine)
	}
	return &restoreRun{data: data, app: app, setting: setting, managed: managed, engine: engine, state: state}, nil
}

func (e *Executor) saveLogs(ctx context.Context, db database.IDB, data *queue.TaskExecData) error {
	frames, err := data.LogStore.GetData(ctx, 0)
	if err != nil {
		return hperrors.Wrap(err)
	}
	_ = data.LogStore.Reset() //nolint:contextcheck // the store resets itself
	return e.saveLogFrames(ctx, db, data.Task, frames)
}

func (e *Executor) saveLogFrames(ctx context.Context, db database.IDB, task *entity.Task,
	frames []*tasklog.LogFrame) error {
	for _, chunk := range gofn.Chunk(frames, 10000) { //nolint:mnd
		taskLogs := make([]*entity.TaskLog, 0, len(chunk))
		for _, frame := range chunk {
			taskLogs = append(taskLogs, &entity.TaskLog{
				TaskID: task.ID, TargetID: task.TargetID, Type: frame.Type, Data: frame.Data, Ts: frame.Ts,
			})
		}
		if err := e.taskLogRepo.InsertMulti(ctx, db, taskLogs); err != nil {
			return hperrors.Wrap(err)
		}
	}
	return nil
}
