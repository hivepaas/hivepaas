package appuc

import (
	"context"
	"fmt"
	"time"

	"github.com/docker/docker/api/types/container"

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

func (uc *UC) GetAppRuntimeLogs(
	ctx context.Context,
	auth *basedto.Auth,
	req *appdto.GetAppRuntimeLogsReq,
) (*appdto.GetAppRuntimeLogsResp, error) {
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

	var since, until, tail string
	if req.Duration > 0 && req.Since.IsZero() {
		req.Since = timeutil.NowUTC().Add(-req.Duration)
		req.Duration = 0
	}
	if !req.Since.IsZero() {
		since = fmt.Sprintf("%d", req.Since.Unix())
	}
	if !req.Follow && req.Duration > 0 {
		until = fmt.Sprintf("%d", req.Since.Add(req.Duration).Unix())
	}
	if req.Tail > 0 {
		tail = fmt.Sprintf("%d", req.Tail)
	}

	logsReader, err := uc.dockerManager.ServiceLogs(ctx, app.ServiceID, func(opts *container.LogsOptions) {
		opts.ShowStdout = true
		opts.ShowStderr = true
		opts.Follow = req.Follow
		opts.Timestamps = req.Timestamps
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

	resp := &appdto.AppRuntimeLogsDataResp{}
	if req.Follow {
		// NOTE: we don't want to keep the log stream session forever
		ctx, _ = context.WithTimeout(ctx, runtimeLogSessionTimeout) //nolint:govet

		// NOTE: We may want to send log frames to client by batch to reduce network overhead.
		// I'm not expert about this, appreciate if anyone can verify this solution.
		// This solution: only send data to client after a period of time or when we have some frames.
		resp.LogChan, resp.LogChanCloser = docker.StartScanningLog(ctx, logsReader,
			docker.WithParseLogHeader(true),
			docker.WithParseLogTimestamp(req.Timestamps),
			docker.WithBatchRecvOptions(batchrecvchan.Options{
				ThresholdPeriod: runtimeLogBatchThresholdPeriod,
				MaxItem:         runtimeLogBatchMaxFrame,
			}),
		)
	} else {
		// Scan all data at once
		logChan, _ := docker.StartScanningLog(ctx, logsReader,
			docker.WithParseLogHeader(true),
			docker.WithParseLogTimestamp(req.Timestamps),
		)
		for frames := range logChan {
			resp.Logs = append(resp.Logs, frames...)
		}
	}

	return &appdto.GetAppRuntimeLogsResp{
		Data: resp,
	}, nil
}
