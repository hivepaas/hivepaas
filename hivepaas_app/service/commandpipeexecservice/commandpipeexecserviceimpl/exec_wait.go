package commandpipeexecserviceimpl

import (
	"context"
	"fmt"
	"time"

	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/tasklog"
)

const (
	taskFindRetryMax           = 10
	taskFindRetryDelay         = time.Second * 3
	taskFindMinRunningDuration = time.Second * 5
)

func (s *service) waitUntilAppsRunning(
	ctx context.Context,
	data *execData,
) error {
	if data.SrcApp != nil && data.SrcApp.ServiceID != "" {
		if err := s.waitUntilAppRunning(ctx, data.SrcApp, data); err != nil {
			return apperrors.Wrap(err)
		}
	}
	if data.DestApp != nil && data.DestApp.ServiceID != "" && (data.SrcApp == nil || data.DestApp.ID != data.SrcApp.ID) {
		if err := s.waitUntilAppRunning(ctx, data.DestApp, data); err != nil {
			return apperrors.Wrap(err)
		}
	}
	return nil
}

func (s *service) waitUntilAppRunning(
	ctx context.Context,
	app *entity.App,
	data *execData,
) error {
	_ = data.LogStore.Add(ctx, tasklog.NewOutFrame(
		fmt.Sprintf("Waiting for app '%s' to be running...", app.Name),
		tasklog.TsNow,
	))

	minRunningDuration := gofn.Coalesce(data.TaskMinRunningDuration, taskFindMinRunningDuration)
	maxRetry := gofn.Coalesce(data.TaskFindRetryMax, taskFindRetryMax)
	retryDelay := gofn.Coalesce(data.TaskFindRetryDelay, taskFindRetryDelay)

	task, _, err := s.dockerManager.ServiceTaskGetRunning(
		ctx,
		app.ServiceID,
		minRunningDuration,
		maxRetry,
		retryDelay,
		nil,
	)
	if err != nil {
		return apperrors.Wrap(err)
	}
	if task == nil {
		_ = data.LogStore.Add(ctx, tasklog.NewWarnFrame(
			fmt.Sprintf("No running task found for app '%s', execution aborted", app.Name),
			tasklog.TsNow,
		))
		return apperrors.NewNotFound(fmt.Sprintf("Running task of app '%s'", app.Name))
	}

	return nil
}
