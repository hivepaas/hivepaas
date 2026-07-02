package appdeploymentserviceimpl

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity/cacheentity"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/tasklog"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/appdeploymentservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/notificationservice"
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
	data.OnPostTransaction(func() { s.onPostTransaction(context.Background(), data) }) //nolint:contextcheck
	s.initLogStore(data)

	err = s.loadDeploymentData(ctx, db, data)
	if err != nil {
		return nil, apperrors.New(err)
	}
	resp.Deployment = data.Deployment

	defer func() {
		if r := recover(); r != nil {
			err = errors.Join(err, apperrors.NewPanic(r))
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
		return apperrors.New(err)
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
		return apperrors.New(err)
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
		return apperrors.New(err)
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
		return apperrors.New(err)
	}
	data.AddRefObjects(refObjects)

	return nil
}

func (s *service) initLogStore(data *appDeploymentData) {
	data.LogStore = tasklog.NewRemoteStore(fmt.Sprintf("task:%s:log", data.Task.ID), s.redisClient)
	data.LogStore.SetOnFlush(tasklog.DefaultMaxSize, func(ctx context.Context, frames []*tasklog.LogFrame) error {
		return s.saveLogFramesToDB(ctx, s.db, data.Task.ID, data.Deployment.ID, frames)
	})
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
		return apperrors.New(err)
	}
	_ = logStore.Close() //nolint

	return s.saveLogFramesToDB(ctx, db, data.Task.ID, deployment.ID, logFrames)
}

func (s *service) saveLogFramesToDB(
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
		err := s.taskLogRepo.InsertMulti(ctx, db, taskLogs)
		if err != nil {
			return apperrors.New(err)
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
