package appdeploymentserviceimpl

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
	"github.com/localpaas/localpaas/localpaas_app/pkg/applog"
	"github.com/localpaas/localpaas/localpaas_app/pkg/bunex"
	"github.com/localpaas/localpaas/localpaas_app/pkg/funcutil"
	"github.com/localpaas/localpaas/localpaas_app/pkg/timeutil"
	"github.com/localpaas/localpaas/localpaas_app/service/appdeploymentservice"
	"github.com/localpaas/localpaas/localpaas_app/service/notificationservice"
)

const (
	deploymentInfoCacheExp = 4 * time.Hour
)

type appDeploymentData struct {
	*appdeploymentservice.AppDeploymentReq
	Project          *entity.Project
	App              *entity.App
	Deployment       *entity.Deployment
	DeploymentOutput *entity.AppDeploymentOutput
	Step             string
	NotifMsgData     *notificationservice.TemplateDataAppDeployment
}

func (s *service) Deploy(
	ctx context.Context,
	db database.Tx,
	req *appdeploymentservice.AppDeploymentReq,
) (resp *appdeploymentservice.AppDeploymentResp, err error) {
	resp = &appdeploymentservice.AppDeploymentResp{}
	data := &appDeploymentData{
		AppDeploymentReq: req,
		DeploymentOutput: &entity.AppDeploymentOutput{},
	}
	logStoreKey := fmt.Sprintf("task:%s:log", req.Task.ID)
	data.LogStore = applog.NewRemoteStore(logStoreKey, true, s.redisClient)
	data.OnPostTransaction = func() { s.onPostTransaction(context.Background(), data) } //nolint:contextcheck

	err = s.loadDeploymentData(ctx, db, data)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	defer func() {
		_ = s.deploymentInfoRepo.Del(ctx, data.Deployment.ID)
		_ = s.saveLogs(ctx, db, data, true)
	}()
	defer funcutil.EnsureNoPanic(&err) // Make sure we catch panic before the above defer

	var depErr error
	depSettings := data.Deployment.Settings
	switch depSettings.ActiveMethod {
	case base.DeploymentMethodImage:
		depErr = s.deployFromImage(ctx, data)
	case base.DeploymentMethodRepo:
		depErr = s.deployFromRepo(ctx, db, data)
	}

	data.Deployment.EndedAt = timeutil.NowUTC()
	if data.Canceled {
		data.Deployment.Status = base.DeploymentStatusCanceled
	} else {
		data.Deployment.Status = gofn.If(depErr != nil, base.DeploymentStatusFailed, base.DeploymentStatusDone)
		data.Deployment.Output = data.DeploymentOutput
	}

	err = s.deploymentRepo.Update(ctx, db, data.Deployment)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	return resp, nil
}

func (s *service) loadDeploymentData(
	ctx context.Context,
	db database.Tx,
	data *appDeploymentData,
) error {
	task := data.Task
	args, err := task.ArgsAsAppDeploy()
	if err != nil {
		return apperrors.Wrap(err)
	}

	deployment, err := s.deploymentRepo.GetByID(ctx, db, "", args.Deployment.ID,
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
	err = s.deploymentInfoRepo.Set(ctx, deployment.ID, &cacheentity.DeploymentInfo{
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

	// Reference setting IDs to load
	refObjectIDs := data.Deployment.Settings.GetRefObjectIDs()

	// Load reference objects
	refObjects, err := s.settingService.LoadReferenceObjectsByIDs(ctx, db, data.App.GetSettingScope(),
		true, true, refObjectIDs)
	if err != nil {
		return apperrors.Wrap(err)
	}
	data.AddRefObjects(refObjects)

	return nil
}

func (s *service) saveLogs(
	ctx context.Context,
	db database.IDB,
	data *appDeploymentData,
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
		err = s.taskLogRepo.InsertMulti(ctx, db, taskLogs)
		if err != nil {
			return apperrors.Wrap(err)
		}
	}

	return nil
}

func (s *service) addStepStartLog(
	ctx context.Context,
	data *appDeploymentData,
	msg string,
) {
	_ = data.LogStore.Add(ctx,
		applog.NewOutFrame("---------------------------------", applog.TsNow),
		applog.NewOutFrame(msg, applog.TsNow))
}

func (s *service) addStepEndLog(
	ctx context.Context,
	data *appDeploymentData,
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

func (s *service) onPostTransaction(
	ctx context.Context,
	data *appDeploymentData,
) {
	db := s.db
	defer func() {
		_ = s.saveLogs(ctx, db, data, false)
	}()

	if data.Task.IsDone() || data.Task.IsFailedCompletely() {
		err := s.notifyForDeployment(ctx, db, data)
		if err != nil {
			_ = data.LogStore.Add(ctx, applog.NewOutFrame("Failed to send deployment notification"+
				" with error: "+err.Error(), applog.TsNow))
		}
	}
}
