package taskserviceimpl

import (
	"context"
	"errors"
	"fmt"

	"github.com/tiendc/gofn"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/entity"
	"github.com/localpaas/localpaas/localpaas_app/infra/database"
	"github.com/localpaas/localpaas/localpaas_app/pkg/batchrecvchan"
	"github.com/localpaas/localpaas/localpaas_app/pkg/bunex"
	"github.com/localpaas/localpaas/localpaas_app/pkg/tasklog"
	"github.com/localpaas/localpaas/localpaas_app/pkg/timeutil"
	"github.com/localpaas/localpaas/localpaas_app/service/taskservice"
	"github.com/localpaas/localpaas/localpaas_app/usecase/system/taskuc/taskdto"
)

func (s *service) GetTaskLogs(
	ctx context.Context,
	db database.IDB,
	req *taskservice.GetTaskLogsReq,
) (*taskservice.GetTaskLogsResp, error) {
	if req.Duration > 0 && req.Since.IsZero() {
		req.Since = timeutil.NowUTC().Add(-req.Duration)
		req.Duration = 0
	}

	task, err := s.taskRepo.GetByID(ctx, db, "", req.TaskID)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	resp := &taskservice.GetTaskLogsResp{}
	err = s.queryLogsInDB(ctx, db, task, req, resp)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	err = s.queryRealtimeLogs(ctx, task, req, resp)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	return resp, nil
}

func (s *service) queryRealtimeLogs(
	ctx context.Context,
	task *entity.Task,
	req *taskservice.GetTaskLogsReq,
	resp *taskservice.GetTaskLogsResp,
) error {
	if task.IsDone() || task.IsCanceled() || task.IsFailedCompletely() {
		return nil
	}

	taskInfo, err := s.taskInfoRepo.Get(ctx, task.ID)
	if err != nil && !errors.Is(err, apperrors.ErrNotFound) {
		return apperrors.Wrap(err)
	}
	if taskInfo == nil {
		return nil
	}

	key := fmt.Sprintf("task:%s:log", task.ID)
	consumer := tasklog.NewConsumer(key, s.redisClient)

	if req.Follow {
		// NOTE: we don't want to keep the log stream session forever
		ctx, _ = context.WithTimeout(ctx, req.LogSessionTimeout) //nolint:govet

		resp.LogsStream, resp.LogsStreamCloser, err = consumer.StartConsuming(ctx, batchrecvchan.Options{
			ThresholdPeriod: req.LogBatchThresholdPeriod,
			MaxItem:         req.LogBatchMaxFrame,
		})
		if err != nil {
			return apperrors.Wrap(err)
		}
	} else {
		frames, err := consumer.GetAllData(ctx)
		if err != nil {
			return apperrors.Wrap(err)
		}
		resp.StaticLogs = append(resp.StaticLogs, frames...)
	}
	return nil
}

func (s *service) queryLogsInDB(
	ctx context.Context,
	db database.IDB,
	task *entity.Task,
	req *taskservice.GetTaskLogsReq,
	resp *taskservice.GetTaskLogsResp,
) error {
	var listOpts []bunex.SelectQueryOption

	reverseLogs := false
	if req.Tail > 0 {
		listOpts = append(listOpts, bunex.SelectLimit(req.Tail),
			bunex.SelectOrder("id DESC"))
		reverseLogs = true
	} else {
		listOpts = append(listOpts, bunex.SelectOrder("id"))
	}

	if !req.Since.IsZero() {
		listOpts = append(listOpts,
			bunex.SelectWhere("task_log.ts >= ?", req.Since))
		if req.Duration > 0 {
			listOpts = append(listOpts,
				bunex.SelectWhere("task_log.ts < ?", req.Since.Add(req.Duration)))
		}
	}

	logs, _, err := s.taskLogRepo.List(ctx, db, task.ID, "", nil, listOpts...)
	if err != nil {
		return apperrors.Wrap(err)
	}

	// Reverse the data
	if reverseLogs {
		gofn.Reverse(logs)
	}

	resp.StaticLogs = append(resp.StaticLogs, taskdto.TransformTaskLogs(logs)...)
	return nil
}
