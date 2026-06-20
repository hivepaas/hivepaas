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
	"github.com/localpaas/localpaas/localpaas_app/pkg/bunex"
	"github.com/localpaas/localpaas/localpaas_app/pkg/tasklog"
	"github.com/localpaas/localpaas/localpaas_app/pkg/timeutil"
	"github.com/localpaas/localpaas/localpaas_app/service/appdeploymentservice"
	"github.com/localpaas/localpaas/localpaas_app/service/notificationservice"
)

const (
	deploymentInfoCacheExp = 4 * time.Hour
)

type appDeploymentData struct {
	*appdeploymentservice.AppDeploymentReq
	Project            *entity.Project
	App                *entity.App
	Deployment         *entity.Deployment
	DeploymentOutput   *entity.AppDeploymentOutput
	DeploymentCanceled bool
	Step               string
	NotifMsgData       *notificationservice.TemplateDataAppDeployment
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
	data.LogStore = tasklog.NewRemoteStore(logStoreKey, true, s.redisClient)
	data.OnPostTransaction(func() { s.onPostTransaction(context.Background(), data) }) //nolint:contextcheck

	err = s.loadDeploymentData(ctx, db, data)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}
	resp.Deployment = data.Deployment

	defer func() {
		if r := recover(); r != nil {
			err = errors.Join(err, apperrors.NewPanic(apperrors.Fmt("%v", r)))
		}
		data.Deployment.UpdatedAt = timeutil.NowUTC()
		data.Deployment.EndedAt = data.Deployment.UpdatedAt
		switch {
		case data.TaskCanceled, data.DeploymentCanceled:
			data.Deployment.Status = base.DeploymentStatusCanceled
		default:
			data.Deployment.Status = gofn.If(err != nil, base.DeploymentStatusFailed, base.DeploymentStatusDone)
			data.Deployment.Output = data.DeploymentOutput
			if err != nil {
				data.Deployment.Output.Error = err.Error()
			}
		}
		err = s.deploymentRepo.Update(ctx, db, data.Deployment)
		// Make cleanup with ignoring errors
		_ = s.deploymentInfoRepo.Del(ctx, data.Deployment.ID)
		_ = s.saveLogs(ctx, db, data, true)
	}()

	switch data.Deployment.Settings.ActiveMethod {
	case base.DeploymentMethodImage:
		err = s.deployFromImage(ctx, db, data)
	case base.DeploymentMethodRepo:
		err = s.deployFromRepo(ctx, db, data)
	}

	return resp, err
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
	refObjects, err := s.settingService.LoadReferenceObjectsByIDs(ctx, db, data.App.GetObjectScope(),
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
		_ = logStore.Add(ctx, tasklog.NewOutFrame("---------------------------------",
			tasklog.TsNow))
		_ = logStore.Add(ctx, tasklog.NewOutFrame("Deployment finished in "+
			deployment.GetDuration().Truncate(time.Millisecond).String(), tasklog.TsNow))
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
		tasklog.NewOutFrame("---------------------------------", tasklog.TsNow),
		tasklog.NewOutFrame(msg, tasklog.TsNow))
}

func (s *service) addStepEndLog(
	ctx context.Context,
	data *appDeploymentData,
	start time.Time,
	err error,
) {
	duration := timeutil.NowUTC().Sub(start).Truncate(time.Millisecond)
	if err != nil {
		_ = data.LogStore.Add(ctx, tasklog.NewOutFrame("Task finished in "+duration.String()+
			" with error: "+err.Error(), tasklog.TsNow))
	} else {
		_ = data.LogStore.Add(ctx, tasklog.NewOutFrame("Task finished in "+duration.String(),
			tasklog.TsNow))
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
			_ = data.LogStore.Add(ctx,
				tasklog.NewOutFrame("---------------------------------", tasklog.TsNow),
				tasklog.NewOutFrame("Failed to send deployment notification with error: "+err.Error(),
					tasklog.TsNow))
		}
	}
}
