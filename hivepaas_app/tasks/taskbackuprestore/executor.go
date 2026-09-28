// Package taskbackuprestore puts a backup snapshot back into an app.
//
// It is a task because a restore takes as long as the snapshot is large, and
// because what it did - an app stopped, a directory moved aside - has to be
// told, step by step, in a log a person can read afterwards.
package taskbackuprestore

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
	"github.com/hivepaas/hivepaas/hivepaas_app/service/backupreposervice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/databackupservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/scopeservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/settingservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/tasks/queue"
)

type Executor struct {
	db          *database.DB
	redisClient rediscache.Client

	settingRepo repository.SettingRepo
	taskLogRepo repository.TaskLogRepo

	appService        appservice.Service
	dataBackupService databackupservice.Service
	scopeService      scopeservice.Service
	settingService    settingservice.Service
}

func NewExecutor(
	taskQueue queue.TaskQueue,
	db *database.DB,
	redisClient rediscache.Client,
	settingRepo repository.SettingRepo,
	taskLogRepo repository.TaskLogRepo,
	appService appservice.Service,
	dataBackupService databackupservice.Service,
	scopeService scopeservice.Service,
	settingService settingservice.Service,
) *Executor {
	e := &Executor{
		db:                db,
		redisClient:       redisClient,
		settingRepo:       settingRepo,
		taskLogRepo:       taskLogRepo,
		appService:        appService,
		dataBackupService: dataBackupService,
		scopeService:      scopeService,
		settingService:    settingService,
	}
	taskQueue.RegisterExecutor(base.TaskTypeBackupRestore, e.execute)
	return e
}

func (e *Executor) execute(
	ctx context.Context,
	db database.Tx,
	task *queue.TaskExecData,
) (err error) {
	// A restore half done is not done again by itself: what it left is for a
	// person to look at.
	task.TaskNonRetryable = true
	task.LogStore = tasklog.NewRemoteStore(fmt.Sprintf("task:%s:log", task.Task.ID), e.redisClient)
	task.LogStore.SetOnFlush(tasklog.DefaultMaxSize, func(ctx context.Context, frames []*tasklog.LogFrame) error {
		return e.saveLogFrames(ctx, e.db, task.Task, frames)
	})
	task.OnPostTx(func() { _ = e.saveLogs(context.Background(), e.db, task) }) //nolint:contextcheck
	defer func() {
		if err != nil {
			_ = task.LogStore.Add(ctx, tasklog.NewErrFrame("Restore failed: "+err.Error()+"\n", tasklog.TsNow))
		}
		_ = task.LogStore.Add(ctx, tasklog.NewOutFrame(fmt.Sprintf("Restore finished in %s\n",
			timeutil.NowUTC().Sub(task.Task.StartedAt).Truncate(time.Millisecond)), tasklog.TsNow))
		_ = e.saveLogs(ctx, db, task)
	}()
	defer safego.RecoverTo(&err)

	args, err := task.Task.ArgsAsBackupRestore()
	if err != nil {
		return hperrors.Wrap(err)
	}
	if args == nil {
		return hperrors.Wrap(hperrors.ErrInternal).WithMsgLog("backup restore task has no arguments")
	}

	// The snapshot may have been deleted since the restore was asked for.
	if _, err = e.settingRepo.GetByID(ctx, db, nil, base.SettingTypeBackupSnapshot, task.Task.TargetID,
		false); err != nil {
		return hperrors.Wrap(err).WithExtraDetail("the snapshot to restore is gone")
	}
	repo, err := e.settingRepo.GetByID(ctx, db, nil, base.SettingTypeBackupRepo, args.RepoID, true)
	if err != nil {
		return hperrors.Wrap(err)
	}
	repoScope, err := e.scopeService.LoadObjectScope(ctx, db, repo.Scope, repo.ObjectID, true)
	if err != nil {
		return hperrors.Wrap(err)
	}
	app, err := e.appService.LoadApp(ctx, db, args.ProjectID, args.AppID, true, true,
		bunex.SelectRelation("Project", bunex.SelectExcludeColumns(entity.ProjectDefaultExcludeColumns...)),
		bunex.SelectRelation("ProjectEnv"),
	)
	if err != nil {
		return hperrors.Wrap(err)
	}
	if args.Command != nil {
		// The command's script, and what its env vars reference.
		if task.RefObjects == nil {
			task.RefObjects = entity.NewRefObjects()
		}
		err = e.settingService.LoadRefObjectsByIDs(ctx, db, &task.RefObjects, app.GetObjectScope(), true,
			args.Command.GetRefObjectIDs())
		if err != nil {
			return hperrors.Wrap(err)
		}
	}

	target := backupreposervice.RepoTarget{Scope: repoScope, RepoSetting: repo}
	return hperrors.Wrap(e.dataBackupService.Restore(ctx, db, restoreReqOf(task, args, app, target)))
}

// restoreReqOf is the restore the task's arguments ask for.
func restoreReqOf(
	task *queue.TaskExecData,
	args *entity.TaskBackupRestoreArgs,
	app *entity.App,
	target backupreposervice.RepoTarget,
) *databackupservice.RestoreReq {
	req := &databackupservice.RestoreReq{TaskExecData: task, Target: target, SnapshotID: args.SnapshotID, App: app}
	if args.Command != nil {
		req.Command = &databackupservice.RestoreCommand{Command: args.Command, FileName: args.FileName}
		return req
	}
	req.Volume = &databackupservice.RestoreVolume{
		VolumeID: args.Volume.ID, Subpath: args.Subpath, SnapshotPath: args.SnapshotPath,
		StopApp: args.StopApp, Mode: args.Mode,
	}
	return req
}

func (e *Executor) saveLogs(ctx context.Context, db database.IDB, task *queue.TaskExecData) error {
	frames, err := task.LogStore.GetData(ctx, 0)
	if err != nil {
		return hperrors.Wrap(err)
	}
	_ = task.LogStore.Reset() //nolint:contextcheck // the store resets itself
	return e.saveLogFrames(ctx, db, task.Task, frames)
}

func (e *Executor) saveLogFrames(
	ctx context.Context,
	db database.IDB,
	task *entity.Task,
	frames []*tasklog.LogFrame,
) error {
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
