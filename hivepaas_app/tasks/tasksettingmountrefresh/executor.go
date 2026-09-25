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
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/logging"
	"github.com/hivepaas/hivepaas/hivepaas_app/repository"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/settingmountservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/tasks/queue"
)

type Executor struct {
	appRepo             repository.AppRepo
	settingMountService settingmountservice.Service
	logger              logging.Logger

	// listPreviews is a seam: a test double of the repository cannot read bunex
	// options.
	listPreviews func(ctx context.Context, db database.IDB, app *entity.App) ([]*entity.App, error)
}

func NewExecutor(
	taskQueue queue.TaskQueue,
	appRepo repository.AppRepo,
	settingMountService settingmountservice.Service,
	logger logging.Logger,
) *Executor {
	e := &Executor{appRepo: appRepo, settingMountService: settingMountService, logger: logger}
	e.listPreviews = e.listPreviewsFromRepo
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
		e.refresh(ctx, db, app, output)
		// A preview mounts its parent's inheritable entries.
		previews, err := e.listPreviews(ctx, db, app)
		if err != nil {
			fail(output, appID, err)
			continue
		}
		for _, preview := range previews {
			e.refresh(ctx, db, preview, output)
		}
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

func (e *Executor) refresh(ctx context.Context, db database.IDB, app *entity.App,
	output *entity.TaskSettingMountRefreshOutput) {
	if err := e.settingMountService.Refresh(ctx, db, app); err != nil {
		e.logger.Errorf("failed to refresh the setting mounts of app %s: %v", app.ID, err)
		fail(output, app.ID, err)
		return
	}
	output.Applied++
}

func (e *Executor) listPreviewsFromRepo(ctx context.Context, db database.IDB, app *entity.App) ([]*entity.App, error) {
	previews, _, err := e.appRepo.List(ctx, db, app.ProjectID, nil, bunex.SelectWhere("app.parent_id = ?", app.ID))
	return previews, hperrors.Wrap(err)
}
