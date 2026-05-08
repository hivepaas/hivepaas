package taskcronjobexec

import (
	"context"
	"fmt"

	"github.com/tiendc/gofn"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/base"
	"github.com/localpaas/localpaas/localpaas_app/entity"
	"github.com/localpaas/localpaas/localpaas_app/infra/database"
	"github.com/localpaas/localpaas/localpaas_app/infra/logging"
	"github.com/localpaas/localpaas/localpaas_app/infra/rediscache"
	"github.com/localpaas/localpaas/localpaas_app/pkg/applog"
	"github.com/localpaas/localpaas/localpaas_app/pkg/funcutil"
	"github.com/localpaas/localpaas/localpaas_app/pkg/timeutil"
	"github.com/localpaas/localpaas/localpaas_app/repository"
	"github.com/localpaas/localpaas/localpaas_app/service/appservice"
	"github.com/localpaas/localpaas/localpaas_app/service/cronjobservice"
	"github.com/localpaas/localpaas/localpaas_app/service/notificationservice"
	"github.com/localpaas/localpaas/localpaas_app/service/settingservice"
	"github.com/localpaas/localpaas/localpaas_app/service/sslservice"
	"github.com/localpaas/localpaas/localpaas_app/service/sysbackupservice"
	"github.com/localpaas/localpaas/localpaas_app/service/syscleanupservice"
	"github.com/localpaas/localpaas/localpaas_app/service/traefikservice"
	"github.com/localpaas/localpaas/localpaas_app/service/userservice"
	"github.com/localpaas/localpaas/localpaas_app/tasks/queue"
	"github.com/localpaas/localpaas/services/docker"
)

type Executor struct {
	logger      logging.Logger
	db          *database.DB
	redisClient rediscache.Client

	userRepo                 repository.UserRepo
	aclPermissionRepo        repository.ACLPermissionRepo
	projectRepo              repository.ProjectRepo
	projectTagRepo           repository.ProjectTagRepo
	projectSharedSettingRepo repository.ProjectSharedSettingRepo
	appRepo                  repository.AppRepo
	appTagRepo               repository.AppTagRepo
	deploymentRepo           repository.DeploymentRepo
	taskLogRepo              repository.TaskLogRepo
	settingRepo              repository.SettingRepo
	taskRepo                 repository.TaskRepo
	sysErrorRepo             repository.SysErrorRepo
	loginTrustedDeviceRepo   repository.LoginTrustedDeviceRepo

	cronJobService      cronjobservice.Service
	appService          appservice.Service
	settingService      settingservice.Service
	sslService          sslservice.Service
	userService         userservice.Service
	notificationService notificationservice.Service
	traefikService      traefikservice.Service
	sysBackupService    sysbackupservice.Service
	sysCleanupService   syscleanupservice.Service
	dockerManager       docker.Manager
}

func NewExecutor(
	logger logging.Logger,
	db *database.DB,
	taskQueue queue.TaskQueue,
	redisClient rediscache.Client,
	userRepo repository.UserRepo,
	aclPermissionRepo repository.ACLPermissionRepo,
	projectRepo repository.ProjectRepo,
	projectTagRepo repository.ProjectTagRepo,
	projectSharedSettingRepo repository.ProjectSharedSettingRepo,
	appRepo repository.AppRepo,
	appTagRepo repository.AppTagRepo,
	deploymentRepo repository.DeploymentRepo,
	taskLogRepo repository.TaskLogRepo,
	settingRepo repository.SettingRepo,
	taskRepo repository.TaskRepo,
	sysErrorRepo repository.SysErrorRepo,
	loginTrustedDeviceRepo repository.LoginTrustedDeviceRepo,
	cronJobService cronjobservice.Service,
	appService appservice.Service,
	settingService settingservice.Service,
	sslService sslservice.Service,
	userService userservice.Service,
	notificationService notificationservice.Service,
	traefikService traefikservice.Service,
	sysBackupService sysbackupservice.Service,
	sysCleanupService syscleanupservice.Service,
	dockerManager docker.Manager,
) *Executor {
	e := &Executor{
		logger:                   logger,
		db:                       db,
		redisClient:              redisClient,
		userRepo:                 userRepo,
		aclPermissionRepo:        aclPermissionRepo,
		projectRepo:              projectRepo,
		projectTagRepo:           projectTagRepo,
		projectSharedSettingRepo: projectSharedSettingRepo,
		appRepo:                  appRepo,
		appTagRepo:               appTagRepo,
		deploymentRepo:           deploymentRepo,
		taskLogRepo:              taskLogRepo,
		settingRepo:              settingRepo,
		taskRepo:                 taskRepo,
		sysErrorRepo:             sysErrorRepo,
		loginTrustedDeviceRepo:   loginTrustedDeviceRepo,
		cronJobService:           cronJobService,
		appService:               appService,
		settingService:           settingService,
		sslService:               sslService,
		userService:              userService,
		notificationService:      notificationService,
		traefikService:           traefikService,
		sysBackupService:         sysBackupService,
		sysCleanupService:        sysCleanupService,
		dockerManager:            dockerManager,
	}
	taskQueue.RegisterExecutor(base.TaskTypeCronJobExec, e.execute)
	return e
}

type taskData struct {
	*queue.TaskExecData
	CronJob      *entity.Setting
	Project      *entity.Project
	App          *entity.App
	NotifMsgData *notificationservice.TemplateDataCronTask
}

func (e *Executor) execute(
	ctx context.Context,
	db database.Tx,
	task *queue.TaskExecData,
) (err error) {
	data := &taskData{
		TaskExecData: task,
		CronJob:      task.Task.TargetJob,
	}
	data.LogStore = applog.NewLocalStore(fmt.Sprintf("cron:%s:exec", data.CronJob.ID))
	data.OnPostTransaction = func() { e.onPostTransaction(data) } //nolint:contextcheck

	err = e.loadCronJobData(ctx, db, data)
	if err != nil {
		return apperrors.Wrap(err)
	}

	defer func() {
		_ = e.saveLogs(ctx, db, data, true)
	}()
	defer funcutil.EnsureNoPanic(&err) // Make sure we catch panic before the above defer

	cronJob := data.CronJob.MustAsCronJob()
	switch cronJob.CronType {
	case base.CronJobTypeContainerCommand:
		err = e.cronExecContainerCmd(ctx, db, data)
	case base.CronJobTypeSystemCleanup:
		_, err = e.sysCleanupService.Cleanup(ctx, db, &syscleanupservice.SysCleanupReq{
			TaskExecData: data.TaskExecData,
			CronJob:      data.CronJob,
		})
	case base.CronJobTypeSystemBackup:
		_, err = e.sysBackupService.Backup(ctx, db, &sysbackupservice.SysBackupReq{
			TaskExecData: data.TaskExecData,
			CronJob:      data.CronJob,
		})
	case base.CronJobTypeSSLRenewal:
		err = e.cronExecSSLRenew(ctx, db, data)
	}
	if err != nil {
		return apperrors.Wrap(err)
	}

	return nil
}

func (e *Executor) loadCronJobData(
	ctx context.Context,
	db database.Tx,
	data *taskData,
) (err error) {
	cronJob := data.CronJob.MustAsCronJob()
	// Load reference objects
	scope := &base.SettingScope{AppID: cronJob.App.ID} // ID can be empty
	refObjects, err := e.settingService.LoadReferenceObjects(ctx, db, scope,
		true, false, data.CronJob)
	if err != nil {
		return apperrors.Wrap(err)
	}
	data.AddRefObjects(refObjects)

	if cronJob.App.ID != "" {
		data.App = data.RefObjects.RefApps[cronJob.App.ID]
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
		_ = logStore.Add(ctx, applog.NewOutFrame("Cron execution finished in "+duration.String(),
			applog.TsNow))
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
				TargetID: data.CronJob.ID,
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
	data *taskData,
) {
	ctx := context.Background()
	db := e.db

	defer func() {
		_ = e.saveLogs(ctx, db, data, false)
	}()

	if data.Task.IsDone() || data.Task.IsFailedCompletely() {
		err := e.sendNotification(ctx, db, data)
		if err != nil {
			_ = data.LogStore.Add(ctx, applog.NewOutFrame("Failed to send result notification"+
				" with error: "+err.Error(), applog.TsNow))
		}
	}
}
