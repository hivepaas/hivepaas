// Package taskimagescan scans the images an app's tasks run: after a
// deployment, or when somebody asks. See service/imagescanservice.
package taskimagescan

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
	"github.com/hivepaas/hivepaas/hivepaas_app/service/imagescanservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/tasks/queue"
)

type Executor struct {
	db          *database.DB
	redisClient rediscache.Client

	appRepo     repository.AppRepo
	taskLogRepo repository.TaskLogRepo

	imageScanService imagescanservice.Service
}

func NewExecutor(
	taskQueue queue.TaskQueue,
	db *database.DB,
	redisClient rediscache.Client,
	appRepo repository.AppRepo,
	taskLogRepo repository.TaskLogRepo,
	imageScanService imagescanservice.Service,
) *Executor {
	e := &Executor{
		db:               db,
		redisClient:      redisClient,
		appRepo:          appRepo,
		taskLogRepo:      taskLogRepo,
		imageScanService: imageScanService,
	}
	taskQueue.RegisterExecutor(base.TaskTypeImageScan, e.execute)
	return e
}

func (e *Executor) execute(
	ctx context.Context,
	db database.Tx,
	task *queue.TaskExecData,
) (err error) {
	task.LogStore = tasklog.NewRemoteStore(fmt.Sprintf("task:%s:log", task.Task.ID), e.redisClient)
	task.LogStore.SetOnFlush(tasklog.DefaultMaxSize, func(ctx context.Context, frames []*tasklog.LogFrame) error {
		return e.saveLogFrames(ctx, e.db, task.Task, frames)
	})
	task.OnPostTx(func() { _ = e.saveLogs(context.Background(), e.db, task) }) //nolint:contextcheck
	defer func() {
		if err != nil {
			_ = task.LogStore.Add(ctx, tasklog.NewErrFrame("Scan failed: "+hperrors.GetErrorDetail(err, "")+"\n",
				tasklog.TsNow))
		}
		_ = task.LogStore.Add(ctx, tasklog.NewOutFrame(fmt.Sprintf("Scan finished in %s\n",
			timeutil.NowUTC().Sub(task.Task.StartedAt).Truncate(time.Millisecond)), tasklog.TsNow))
		_ = e.saveLogs(ctx, db, task)
	}()
	defer safego.RecoverTo(&err)

	args, err := task.Task.ArgsAsImageScan()
	if err != nil {
		return hperrors.Wrap(err)
	}
	if args == nil || args.AppID == "" {
		return hperrors.Wrap(hperrors.ErrInternal).WithMsgLog("image scan task has no app")
	}
	app, err := e.appRepo.GetByID(ctx, db, "", args.AppID,
		bunex.SelectRelation("Project", bunex.SelectExcludeColumns(entity.ProjectDefaultExcludeColumns...)),
		bunex.SelectRelation("ProjectEnv"),
	)
	if err != nil {
		return hperrors.Wrap(err)
	}

	output, err := e.imageScanService.ScanApp(ctx, &imagescanservice.ScanAppReq{
		App:      app,
		Trigger:  args.Trigger,
		Force:    args.Force,
		LogStore: task.LogStore,
	})
	if output != nil {
		task.Task.MustSetOutput(output)
	}
	return hperrors.Wrap(err)
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
