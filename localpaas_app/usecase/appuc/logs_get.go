package appuc

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"time"

	"github.com/moby/moby/client"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/entity"
	"github.com/localpaas/localpaas/localpaas_app/pkg/batchrecvchan"
	"github.com/localpaas/localpaas/localpaas_app/pkg/bunex"
	"github.com/localpaas/localpaas/localpaas_app/pkg/timeutil"
	"github.com/localpaas/localpaas/localpaas_app/usecase/appuc/appdto"
	"github.com/localpaas/localpaas/services/docker"
)

const (
	runtimeLogBatchThresholdPeriod = time.Millisecond * 500
	runtimeLogBatchMaxFrame        = 20
	runtimeLogSessionTimeout       = time.Hour
)

//nolint:gocognit,funlen
func (uc *UC) GetAppLogs(
	ctx context.Context,
	auth *basedto.Auth, // BE CAREFUL: If req.Token presents, auth is nil
	req *appdto.GetAppLogsReq,
) (_ *appdto.GetAppLogsResp, err error) {
	var serviceID string
	if auth == nil { //nolint:nestif
		req.Token, err = url.QueryUnescape(req.Token)
		if err != nil {
			return nil, apperrors.New(apperrors.ErrTokenInvalid).WithCause(err)
		}
		ticketInfo, err := uc.consoleTicketRepo.Get(ctx, req.Token)
		if err != nil {
			return nil, apperrors.New(apperrors.ErrTokenInvalid).WithCause(err)
		}
		if req.AppID != ticketInfo.AppID {
			return nil, apperrors.New(apperrors.ErrTokenInvalid)
		}
		serviceID = ticketInfo.TargetID
		// Remove the ticket from redis as this ticket is one-time object
		_ = uc.consoleTicketRepo.Del(ctx, req.Token)
	} else {
		app, err := uc.appRepo.GetByID(ctx, uc.db, req.ProjectID, req.AppID,
			bunex.SelectExcludeColumns(entity.AppDefaultExcludeColumns...),
		)
		if err != nil {
			return nil, apperrors.Wrap(err)
		}
		if app.ServiceID == "" {
			return nil, apperrors.New(apperrors.ErrUnavailable).
				WithMsgLog("service not exist for app")
		}
		serviceID = app.ServiceID
	}

	var since, until, tail string
	if req.Duration > 0 && req.Since.IsZero() {
		req.Since = timeutil.NowUTC().Add(-req.Duration.ToDuration())
		req.Duration = 0
	}
	if !req.Since.IsZero() {
		since = fmt.Sprintf("%d", req.Since.Unix())
	}
	if !req.Follow && req.Duration > 0 {
		until = fmt.Sprintf("%d", req.Since.Add(req.Duration.ToDuration()).Unix())
	}
	if req.Tail > 0 {
		tail = fmt.Sprintf("%d", req.Tail)
	}
	if req.Timestamps == nil {
		req.Timestamps = new(true)
	}

	var logsReader io.ReadCloser
	if req.TaskID != "" { //nolint:nestif
		// Validate task belongs to the service
		taskInspect, err := uc.dockerManager.TaskInspect(ctx, req.TaskID)
		if err != nil {
			return nil, apperrors.Wrap(err)
		}
		if taskInspect.Task.ServiceID != serviceID {
			return nil, apperrors.New(apperrors.ErrUnavailable).
				WithMsgLog("task doesn't belong to service")
		}

		logsReader, err = uc.dockerManager.TaskLogs(ctx, req.TaskID, func(opts *client.TaskLogsOptions) {
			opts.ShowStdout = true
			opts.ShowStderr = true
			opts.Follow = req.Follow
			opts.Timestamps = *req.Timestamps
			if since != "" {
				opts.Since = since
			}
			if until != "" {
				opts.Until = until
			}
			if tail != "" {
				opts.Tail = tail
			}
		})
		if err != nil {
			return nil, apperrors.Wrap(err)
		}
	} else {
		logsReader, err = uc.dockerManager.ServiceLogs(ctx, serviceID, func(opts *client.ServiceLogsOptions) {
			opts.ShowStdout = true
			opts.ShowStderr = true
			opts.Follow = req.Follow
			opts.Timestamps = *req.Timestamps
			if since != "" {
				opts.Since = since
			}
			if until != "" {
				opts.Until = until
			}
			if tail != "" {
				opts.Tail = tail
			}
		})
		if err != nil {
			return nil, apperrors.Wrap(err)
		}
	}

	resp := &appdto.AppLogsDataResp{}
	if req.Follow {
		// NOTE: we don't want to keep the log stream session forever
		ctx, _ = context.WithTimeout(ctx, runtimeLogSessionTimeout) //nolint:govet

		// NOTE: We may want to send log frames to client by batch to reduce network overhead.
		// I'm not an expert about this, appreciate if anyone can verify this solution.
		// This solution: only send data to client after a period of time or when we have some frames.
		resp.LogsStream, resp.LogsStreamCloser = docker.StartScanningLog(ctx, logsReader,
			docker.WithParseLogHeader(true),
			docker.WithParseLogTimestamp(*req.Timestamps),
			docker.WithBatchRecvOptions(batchrecvchan.Options{
				ThresholdPeriod: runtimeLogBatchThresholdPeriod,
				MaxItem:         runtimeLogBatchMaxFrame,
			}),
		)
	} else {
		// Scan all data at once
		logStream, _ := docker.StartScanningLog(ctx, logsReader,
			docker.WithParseLogHeader(true),
			docker.WithParseLogTimestamp(*req.Timestamps),
		)
		for frames := range logStream {
			resp.StaticLogs = append(resp.StaticLogs, frames...)
		}
	}

	return &appdto.GetAppLogsResp{
		Data: resp,
	}, nil
}
