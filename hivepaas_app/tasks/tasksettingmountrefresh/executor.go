// Package tasksettingmountrefresh brings the files apps mount from settings up
// to date after one of those settings changed: a certificate renewed, a password
// replaced, an entry disabled.
package tasksettingmountrefresh

import (
	"context"
	"errors"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/logging"
	"github.com/hivepaas/hivepaas/hivepaas_app/repository"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/settingmountservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/tasks/queue"
)

type Executor struct {
	appRepo             repository.AppRepo
	settingMountService settingmountservice.Service
	logger              logging.Logger
}

func NewExecutor(
	taskQueue queue.TaskQueue,
	appRepo repository.AppRepo,
	settingMountService settingmountservice.Service,
	logger logging.Logger,
) *Executor {
	e := &Executor{appRepo: appRepo, settingMountService: settingMountService, logger: logger}
	taskQueue.RegisterExecutor(base.TaskTypeSettingMountRefresh, e.execute)
	return e
}

func (e *Executor) execute(ctx context.Context, db database.Tx, task *queue.TaskExecData) error {
	args, err := task.Task.ArgsAsSettingMountRefresh()
	if err != nil {
		return hperrors.Wrap(err)
	}
	if args == nil {
		return hperrors.Wrap(hperrors.ErrInternal).WithMsgLog("setting mount refresh task has no args")
	}

	output := &entity.TaskSettingMountRefreshOutput{}
	for _, appID := range args.AppIDs {
		app, err := e.appRepo.GetByID(ctx, db, "", appID)
		if err != nil {
			if errors.Is(err, hperrors.ErrNotFound) {
				continue // deleted since; its objects went with it
			}
			fail(output, appID, err)
			continue
		}
		if err = e.settingMountService.Refresh(ctx, db, app); err != nil {
			e.logger.Errorf("failed to refresh the setting mounts of app %s: %v", appID, err)
			fail(output, appID, err)
			continue
		}
		output.Applied++
	}
	task.Task.MustSetOutput(output)

	// An app that failed is retried with the task; the others cost one inspect
	// each and no update, as nothing of theirs changed.
	if len(output.Failed) > 0 {
		return hperrors.Wrap(hperrors.ErrInternal).
			WithMsgLog("setting mounts could not be refreshed on %d app(s)", len(output.Failed))
	}
	return nil
}

func fail(out *entity.TaskSettingMountRefreshOutput, appID string, err error) {
	if out.Failed == nil {
		out.Failed = map[string]string{}
	}
	out.Failed[appID] = err.Error()
}
