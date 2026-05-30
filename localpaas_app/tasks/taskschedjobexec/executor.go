package taskschedjobexec

import (
	"context"
	"fmt"

	"github.com/tiendc/gofn"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/base"
	"github.com/localpaas/localpaas/localpaas_app/entity"
	"github.com/localpaas/localpaas/localpaas_app/infra/database"
	"github.com/localpaas/localpaas/localpaas_app/infra/logging"
	"github.com/localpaas/localpaas/localpaas_app/pkg/funcutil"
	"github.com/localpaas/localpaas/localpaas_app/pkg/tasklog"
	"github.com/localpaas/localpaas/localpaas_app/pkg/timeutil"
	"github.com/localpaas/localpaas/localpaas_app/repository"
	"github.com/localpaas/localpaas/localpaas_app/service/containerexecservice"
	"github.com/localpaas/localpaas/localpaas_app/service/notificationservice"
	"github.com/localpaas/localpaas/localpaas_app/service/settingservice"
	"github.com/localpaas/localpaas/localpaas_app/service/sslrenewalservice"
	"github.com/localpaas/localpaas/localpaas_app/service/sysbackupservice"
	"github.com/localpaas/localpaas/localpaas_app/service/syscleanupservice"
	"github.com/localpaas/localpaas/localpaas_app/tasks/queue"
)

type Executor struct {
	logger               logging.Logger
	db                   *database.DB
	taskLogRepo          repository.TaskLogRepo
	settingService       settingservice.Service
	notificationService  notificationservice.Service
	containerExecService containerexecservice.Service
	sysBackupService     sysbackupservice.Service
	sysCleanupService    syscleanupservice.Service
	sslRenewalService    sslrenewalservice.Service
}

func NewExecutor(
	logger logging.Logger,
	db *database.DB,
	taskQueue queue.TaskQueue,
	taskLogRepo repository.TaskLogRepo,
	settingService settingservice.Service,
	notificationService notificationservice.Service,
	containerExecService containerexecservice.Service,
	sysBackupService sysbackupservice.Service,
	sysCleanupService syscleanupservice.Service,
	sslRenewalService sslrenewalservice.Service,
) *Executor {
	e := &Executor{
		logger:               logger,
		db:                   db,
		taskLogRepo:          taskLogRepo,
		settingService:       settingService,
		notificationService:  notificationService,
		containerExecService: containerExecService,
		sysBackupService:     sysBackupService,
		sysCleanupService:    sysCleanupService,
		sslRenewalService:    sslRenewalService,
	}
	taskQueue.RegisterExecutor(base.TaskTypeSchedJobExec, e.execute)
	return e
}

type taskData struct {
	*queue.TaskExecData
	SchedJob *entity.Setting
	Project  *entity.Project
	App      *entity.App

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
	data.LogStore = tasklog.NewLocalStore(fmt.Sprintf("job:%s:exec", data.SchedJob.ID))
	data.OnPostTransaction(func() { e.onPostTransaction(context.Background(), data) }) //nolint:contextcheck

	err = e.loadSchedJobData(ctx, db, data)
	if err != nil {
		return apperrors.Wrap(err)
	}

	defer func() {
		_ = e.saveLogs(ctx, db, data, true)
	}()
	defer funcutil.EnsureNoPanic(&err) // Make sure we catch panic before the above defer

	schedJob := data.SchedJob.MustAsSchedJob()
	switch schedJob.JobType {
	case base.SchedJobTypeContainerCommand:
		resp, err := e.containerExecService.ContainerExec(ctx, db, &containerexecservice.ContainerExecReq{
			TaskExecData: data.TaskExecData,
			SchedJob:     data.SchedJob,
			Project:      data.Project,
			App:          data.App,
		})
		if err != nil {
			return apperrors.Wrap(err)
		}
		data.SkipResultNotification = resp.SkipResultNotification

	case base.SchedJobTypeSystemCleanup:
		setting := data.RefObjects.RefSettings[schedJob.TargetSetting.ID]
		if setting == nil {
			return apperrors.NewNotFound("System cleanup settings")
		}
		resp, err := e.sysCleanupService.Cleanup(ctx, db, &syscleanupservice.SysCleanupReq{
			TaskExecData:       data.TaskExecData,
			SysCleanupSettings: setting.MustAsSystemCleanup(),
		})
		if err != nil {
			return apperrors.Wrap(err)
		}
		data.SkipResultNotification = resp.SkipResultNotification

	case base.SchedJobTypeSystemBackup:
		setting := data.RefObjects.RefSettings[schedJob.TargetSetting.ID]
		if setting == nil {
			return apperrors.NewNotFound("System backup settings")
		}
		resp, err := e.sysBackupService.Backup(ctx, db, &sysbackupservice.SysBackupReq{
			TaskExecData:      data.TaskExecData,
			SysBackupSettings: setting.MustAsSystemBackup(),
		})
		if err != nil {
			return apperrors.Wrap(err)
		}
		data.SkipResultNotification = resp.SkipResultNotification

	case base.SchedJobTypeSSLRenewal:
		setting := data.RefObjects.RefSettings[schedJob.TargetSetting.ID]
		if setting == nil {
			return apperrors.NewNotFound("SSL renewal settings")
		}
		resp, err := e.sslRenewalService.SSLRenew(ctx, db, &sslrenewalservice.SSLRenewalReq{
			TaskExecData:    data.TaskExecData,
			SchedJob:        data.SchedJob,
			RenewalSettings: setting.MustAsSSLRenewal(),
		})
		if err != nil {
			return apperrors.Wrap(err)
		}
		data.SkipResultNotification = resp.SkipResultNotification
	}

	return nil
}

func (e *Executor) loadSchedJobData(
	ctx context.Context,
	db database.IDB,
	data *taskData,
) (err error) {
	schedJob := data.SchedJob.MustAsSchedJob()
	// Load reference objects
	scope := &base.ObjectScope{AppID: schedJob.App.ID} // ID can be empty
	refObjects, err := e.settingService.LoadReferenceObjects(ctx, db, scope,
		true, false, data.SchedJob)
	if err != nil {
		return apperrors.Wrap(err)
	}
	data.AddRefObjects(refObjects)

	if schedJob.App.ID != "" {
		data.App = data.RefObjects.RefApps[schedJob.App.ID]
		data.Project = data.App.Project
	}

	return nil
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
		_ = logStore.Add(ctx, tasklog.NewOutFrame("Job execution finished in "+duration.String(),
			tasklog.TsNow))
	}

	logFrames, err := logStore.GetData(ctx, 0)
	if err != nil {
		return apperrors.Wrap(err)
	}
	_ = logStore.Reset() //nolint

	// Insert data in to DB by chunk to avoid exceeding DBMS limit
	for _, chunk := range gofn.Chunk(logFrames, 10000) { //nolint
		taskLogs := make([]*entity.TaskLog, 0, len(chunk))
		for _, logFrame := range chunk {
			taskLogs = append(taskLogs, &entity.TaskLog{
				TaskID:   data.Task.ID,
				TargetID: data.SchedJob.ID,
				Type:     logFrame.Type,
				Data:     logFrame.Data,
				Ts:       logFrame.Ts,
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
			_ = data.LogStore.Add(ctx, tasklog.NewOutFrame("Failed to send result notification"+
				" with error: "+err.Error(), tasklog.TsNow))
		}
	}
}
