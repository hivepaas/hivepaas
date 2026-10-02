package taskschedjobexec

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
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/logging"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/safego"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/tasklog"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/repository"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/appservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/backuprepocleanupservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/databackupservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/notificationservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/registryauthservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/schedjobexecservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/scopeservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/settingservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/sslrenewalservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/sysbackupservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/syscleanupservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/tasks/queue"
)

type Executor struct {
	logger      logging.Logger
	db          *database.DB
	redisClient rediscache.Client

	taskLogRepo repository.TaskLogRepo
	taskRepo    repository.TaskRepo

	appService               appservice.Service
	backupRepoCleanupService backuprepocleanupservice.Service
	dataBackupService        databackupservice.Service
	notificationService      notificationservice.Service
	registryAuthService      registryauthservice.Service
	schedJobExecService      schedjobexecservice.Service
	scopeService             scopeservice.Service
	settingService           settingservice.Service
	sslRenewalService        sslrenewalservice.Service
	sysBackupService         sysbackupservice.Service
	sysCleanupService        syscleanupservice.Service
}

func NewExecutor(
	logger logging.Logger,
	db *database.DB,
	redisClient rediscache.Client,
	taskQueue queue.TaskQueue,

	taskLogRepo repository.TaskLogRepo,
	taskRepo repository.TaskRepo,

	appService appservice.Service,
	backupRepoCleanupService backuprepocleanupservice.Service,
	dataBackupService databackupservice.Service,
	notificationService notificationservice.Service,
	registryAuthService registryauthservice.Service,
	schedJobExecService schedjobexecservice.Service,
	scopeService scopeservice.Service,
	settingService settingservice.Service,
	sslRenewalService sslrenewalservice.Service,
	sysBackupService sysbackupservice.Service,
	sysCleanupService syscleanupservice.Service,
) *Executor {
	e := &Executor{
		logger:      logger,
		db:          db,
		redisClient: redisClient,

		taskLogRepo: taskLogRepo,
		taskRepo:    taskRepo,

		appService:               appService,
		backupRepoCleanupService: backupRepoCleanupService,
		dataBackupService:        dataBackupService,
		notificationService:      notificationService,
		registryAuthService:      registryAuthService,
		schedJobExecService:      schedJobExecService,
		scopeService:             scopeService,
		settingService:           settingService,
		sslRenewalService:        sslRenewalService,
		sysBackupService:         sysBackupService,
		sysCleanupService:        sysCleanupService,
	}
	taskQueue.RegisterExecutor(base.TaskTypeSchedJobExec, e.execute)
	return e
}

type taskData struct {
	*queue.TaskExecData
	SchedJob *entity.Setting
	Scope    *entity.ObjectScope

	SkipResultNotification bool
	NotifMsgData           *notificationservice.TemplateDataSchedTask
}

func (e *Executor) execute(
	ctx context.Context,
	db database.Tx,
	task *queue.TaskExecData,
) (err error) {
	data := &taskData{
		TaskExecData: task,
		SchedJob:     task.Task.TargetJob,
	}
	data.OnPostTx(func() { e.onPostTx(context.Background(), data) }) //nolint:contextcheck
	e.initLogStore(data)

	err = e.loadSchedJobData(ctx, db, data)
	if err != nil {
		return hperrors.Wrap(err)
	}

	defer func() {
		_ = e.saveLogs(ctx, db, data, true)
	}()
	defer safego.RecoverTo(&err) // Make sure we catch panic before the above defer

	if data.SchedJob.MustAsSchedJob().JobType == base.SchedJobTypeJobSequence {
		return e.executeSequence(ctx, db, data)
	}
	result, err := e.runJob(ctx, db, &jobRun{
		execData:   data.TaskExecData,
		jobSetting: data.SchedJob,
		refObjects: data.RefObjects,
	})
	if err != nil {
		return hperrors.Wrap(err)
	}
	data.SkipResultNotification = result.skipNotification
	return nil
}

func (e *Executor) loadSchedJobData(
	ctx context.Context,
	db database.IDB,
	data *taskData,
) (err error) {
	scope, err := e.scopeService.LoadObjectScope(ctx, db, data.SchedJob.Scope, data.SchedJob.ObjectID, true)
	if err != nil {
		return hperrors.Wrap(err)
	}
	data.Scope = scope

	if data.RefObjects == nil {
		data.RefObjects = entity.NewRefObjects()
	}
	data.RefObjects.AddObjectScope(scope)

	// Load reference objects
	err = e.settingService.LoadRefObjectsSkipMissing(ctx, db, &data.RefObjects, scope, true, data.SchedJob)
	if err != nil {
		return hperrors.Wrap(err)
	}

	return nil
}

func (e *Executor) initLogStore(data *taskData) {
	data.LogStore = tasklog.NewRemoteStore(fmt.Sprintf("task:%s:log", data.Task.ID), e.redisClient)
	data.LogStore.SetOnFlush(tasklog.DefaultMaxSize, func(ctx context.Context, frames []*tasklog.LogFrame) error {
		return e.saveLogFramesToDB(ctx, e.db, data.Task.ID, data.SchedJob.ID, frames)
	})
}

func (e *Executor) saveLogs(
	ctx context.Context,
	db database.IDB,
	data *taskData,
	addDurationInfo bool,
) error {
	logStore := data.LogStore
	if logStore == nil {
		return nil
	}

	if addDurationInfo {
		duration := timeutil.NowUTC().Sub(data.Task.StartedAt)
		_ = logStore.Add(ctx,
			tasklog.NewOutFrame("\n---------------------------------\n", tasklog.TsNow),
			tasklog.NewOutFrame("Job execution finished in "+
				duration.Truncate(time.Millisecond).String()+"\n", tasklog.TsNow))
	}

	logFrames, err := logStore.GetData(ctx, 0)
	if err != nil {
		return hperrors.Wrap(err)
	}
	_ = logStore.Reset() //nolint

	return e.saveLogFramesToDB(ctx, db, data.Task.ID, data.SchedJob.ID, logFrames)
}

func (e *Executor) saveLogFramesToDB(
	ctx context.Context,
	db database.IDB,
	taskID string,
	targetID string,
	logFrames []*tasklog.LogFrame,
) error {
	for _, chunk := range gofn.Chunk(logFrames, 10000) { //nolint
		taskLogs := make([]*entity.TaskLog, 0, len(chunk))
		for _, logFrame := range chunk {
			taskLogs = append(taskLogs, &entity.TaskLog{
				TaskID:   taskID,
				TargetID: targetID,
				Type:     logFrame.Type,
				Data:     logFrame.Data,
				Ts:       logFrame.Ts,
			})
		}
		err := e.taskLogRepo.InsertMulti(ctx, db, taskLogs)
		if err != nil {
			return hperrors.Wrap(err)
		}
	}
	return nil
}

func (e *Executor) onPostTx(
	ctx context.Context,
	data *taskData,
) {
	db := e.db
	defer func() {
		_ = e.saveLogs(ctx, db, data, false)
	}()

	if !data.SkipResultNotification && (data.Task.IsDone() || data.Task.IsFailedCompletely()) {
		err := e.sendNotification(ctx, db, data)
		if err != nil {
			_ = data.LogStore.Add(ctx,
				tasklog.NewOutFrame("Failed to send result notification with error: "+err.Error()+"\n",
					tasklog.TsNow))
		}
	}
}
