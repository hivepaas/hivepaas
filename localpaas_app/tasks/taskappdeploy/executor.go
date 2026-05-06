package taskappdeploy

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/tiendc/gofn"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/base"
	"github.com/localpaas/localpaas/localpaas_app/entity"
	"github.com/localpaas/localpaas/localpaas_app/entity/cacheentity"
	"github.com/localpaas/localpaas/localpaas_app/infra/database"
	"github.com/localpaas/localpaas/localpaas_app/infra/logging"
	"github.com/localpaas/localpaas/localpaas_app/infra/rediscache"
	"github.com/localpaas/localpaas/localpaas_app/pkg/applog"
	"github.com/localpaas/localpaas/localpaas_app/pkg/bunex"
	"github.com/localpaas/localpaas/localpaas_app/pkg/funcutil"
	"github.com/localpaas/localpaas/localpaas_app/pkg/timeutil"
	"github.com/localpaas/localpaas/localpaas_app/repository"
	"github.com/localpaas/localpaas/localpaas_app/repository/cacherepository"
	"github.com/localpaas/localpaas/localpaas_app/service/envvarservice"
	"github.com/localpaas/localpaas/localpaas_app/service/notificationservice"
	"github.com/localpaas/localpaas/localpaas_app/service/settingservice"
	"github.com/localpaas/localpaas/localpaas_app/service/userservice"
	"github.com/localpaas/localpaas/localpaas_app/tasks/queue"
	"github.com/localpaas/localpaas/services/docker"
)

const (
	deploymentInfoCacheExp = 4 * time.Hour
)

type Executor struct {
	logger              logging.Logger
	db                  *database.DB
	redisClient         rediscache.Client
	settingRepo         repository.SettingRepo
	deploymentRepo      repository.DeploymentRepo
	taskLogRepo         repository.TaskLogRepo
	taskRepo            repository.TaskRepo
	taskInfoRepo        cacherepository.TaskInfoRepo
	deploymentInfoRepo  cacherepository.DeploymentInfoRepo
	dockerManager       docker.Manager
	envVarService       envvarservice.Service
	settingService      settingservice.Service
	userService         userservice.Service
	notificationService notificationservice.Service
}

func NewExecutor(
	taskQueue queue.TaskQueue,
	logger logging.Logger,
	db *database.DB,
	redisClient rediscache.Client,
	settingRepo repository.SettingRepo,
	deploymentRepo repository.DeploymentRepo,
	taskLogRepo repository.TaskLogRepo,
	taskRepo repository.TaskRepo,
	taskInfoRepo cacherepository.TaskInfoRepo,
	deploymentInfoRepo cacherepository.DeploymentInfoRepo,
	dockerManager docker.Manager,
	envVarService envvarservice.Service,
	settingService settingservice.Service,
	userService userservice.Service,
	notificationService notificationservice.Service,
) *Executor {
	e := &Executor{
		logger:              logger,
		db:                  db,
		redisClient:         redisClient,
		settingRepo:         settingRepo,
		deploymentRepo:      deploymentRepo,
		taskLogRepo:         taskLogRepo,
		taskRepo:            taskRepo,
		taskInfoRepo:        taskInfoRepo,
		deploymentInfoRepo:  deploymentInfoRepo,
		dockerManager:       dockerManager,
		envVarService:       envVarService,
		settingService:      settingService,
		userService:         userService,
		notificationService: notificationService,
	}
	taskQueue.RegisterExecutor(base.TaskTypeAppDeploy, e.execute)
	return e
}

type taskData struct {
	*queue.TaskExecData
	Project          *entity.Project
	App              *entity.App
	Deployment       *entity.Deployment
	DeploymentOutput *entity.AppDeploymentOutput
	Step             string
	LogStore         *applog.Store
	NotifMsgData     *notificationservice.BaseMsgDataAppDeploymentNotification
}

func (e *Executor) execute(
	ctx context.Context,
	db database.Tx,
	task *queue.TaskExecData,
) (err error) {
	data := &taskData{TaskExecData: task}
	data.SetOnPostTransaction(func() { e.onPostTransaction(data) }) //nolint

	err = e.loadDeploymentData(ctx, db, data)
	if err != nil {
		return apperrors.Wrap(err)
	}

	defer func() {
		_ = e.deploymentInfoRepo.Del(ctx, data.Deployment.ID)
		_ = e.saveLogs(ctx, db, data, true)
	}()
	defer funcutil.EnsureNoPanic(&err) // Make sure we catch panic before the above defer

	var depErr error
	depSettings := data.Deployment.Settings
	switch depSettings.ActiveMethod {
	case base.DeploymentMethodImage:
		depErr = e.deployFromImage(ctx, data)
	case base.DeploymentMethodRepo:
		depErr = e.deployFromRepo(ctx, db, data)
	}

	data.Deployment.EndedAt = timeutil.NowUTC()
	if data.Canceled {
		data.Deployment.Status = base.DeploymentStatusCanceled
	} else {
		data.Deployment.Status = gofn.If(depErr != nil, base.DeploymentStatusFailed, base.DeploymentStatusDone)
		data.Deployment.Output = data.DeploymentOutput
	}

	err = e.updateDeployment(ctx, db, data.Deployment)
	if err != nil {
		return apperrors.Wrap(err)
	}

	return nil
}

func (e *Executor) loadDeploymentData(
	ctx context.Context,
	db database.Tx,
	data *taskData,
) error {
	task := data.Task
	args, err := task.ArgsAsAppDeploy()
	if err != nil {
		return apperrors.Wrap(err)
	}

	deployment, err := e.deploymentRepo.GetByID(ctx, db, "", args.Deployment.ID,
		bunex.SelectWhereIn("deployment.status IN (?)",
			base.DeploymentStatusNotStarted, base.DeploymentStatusInProgress),
		bunex.SelectRelation("App",
			bunex.SelectExcludeColumns(entity.AppDefaultExcludeColumns...),
			bunex.SelectWhere("app.status = ?", base.AppStatusActive),
		),
		bunex.SelectRelation("App.Project",
			bunex.SelectExcludeColumns(entity.ProjectDefaultExcludeColumns...),
			bunex.SelectWhere("app__project.status = ?", base.ProjectStatusActive),
		),
		bunex.SelectFor("UPDATE OF deployment"),
	)
	if err != nil && !errors.Is(err, apperrors.ErrNotFound) {
		return apperrors.Wrap(err)
	}
	if deployment == nil || deployment.App == nil || deployment.App.Project == nil { // no active deployment, return
		return nil
	}

	if deployment.Status == base.DeploymentStatusNotStarted {
		deployment.StartedAt = data.Task.StartedAt
		deployment.Status = base.DeploymentStatusInProgress
	}

	// Put deployment status in redis
	err = e.deploymentInfoRepo.Set(ctx, deployment.ID, &cacheentity.DeploymentInfo{
		ID:        deployment.ID,
		AppID:     deployment.AppID,
		TaskID:    task.ID,
		Status:    base.DeploymentStatusInProgress,
		StartedAt: deployment.StartedAt,
	}, deploymentInfoCacheExp)
	if err != nil {
		return apperrors.Wrap(err)
	}

	data.App = deployment.App
	data.Project = data.App.Project
	data.Deployment = deployment
	data.DeploymentOutput = &entity.AppDeploymentOutput{}
	logStoreKey := fmt.Sprintf("task:%s:log", task.ID)
	data.LogStore = applog.NewRemoteStore(logStoreKey, true, e.redisClient)

	// Reference setting IDs to load
	refObjectIDs := data.Deployment.Settings.GetRefObjectIDs()

	// Load reference objects
	refObjects, err := e.settingService.LoadReferenceObjectsByIDs(ctx, db, data.App.GetSettingScope(),
		true, true, refObjectIDs)
	if err != nil {
		return apperrors.Wrap(err)
	}
	data.AddRefObjects(refObjects)

	return nil
}

func (e *Executor) updateDeployment(
	ctx context.Context,
	db database.Tx,
	deployment *entity.Deployment,
) error {
	err := e.deploymentRepo.Update(ctx, db, deployment)
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
	deployment := data.Deployment
	logStore := data.LogStore
	if logStore == nil {
		return nil
	}

	if addDurationInfo {
		_ = logStore.Add(ctx, applog.NewOutFrame("Deployment finished in "+
			deployment.GetDuration().String(), applog.TsNow))
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
				TaskID:   data.Task.ID,
				TargetID: deployment.ID,
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

func (e *Executor) addStepStartLog(
	ctx context.Context,
	data *taskData,
	msg string,
) {
	_ = data.LogStore.Add(ctx,
		applog.NewOutFrame("---------------------------------", applog.TsNow),
		applog.NewOutFrame(msg, applog.TsNow))
}

func (e *Executor) addStepEndLog(
	ctx context.Context,
	data *taskData,
	start time.Time,
	err error,
) {
	duration := timeutil.NowUTC().Sub(start)
	if err != nil {
		_ = data.LogStore.Add(ctx, applog.NewOutFrame("Task finished in "+duration.String()+
			" with error: "+err.Error(), applog.TsNow))
	} else {
		_ = data.LogStore.Add(ctx, applog.NewOutFrame("Task finished in "+duration.String(),
			applog.TsNow))
	}
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
		err := e.notifyForDeployment(ctx, db, data)
		if err != nil {
			_ = data.LogStore.Add(ctx, applog.NewOutFrame("Failed to send deployment notification"+
				" with error: "+err.Error(), applog.TsNow))
		}
	}
}
