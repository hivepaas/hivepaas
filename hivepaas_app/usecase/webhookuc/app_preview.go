package webhookuc

import (
	"context"
	"strconv"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/transaction"
)

func (uc *UC) createAppPreview(
	ctx context.Context,
	app *entity.App,
	commentEvent *repoPRCommentEventData,
	repoRef string,
	webhookID string,
	previewSettings *entity.AppFeaturePreviewSettings, // if nil, will be loaded from DB
) (err error) {
	if app.IsChildApp() { // The app is already a preview app, skips it
		return nil
	}

	var previewTask *entity.Task
	err = transaction.Execute(ctx, uc.db, func(db database.Tx) (err error) {
		if previewSettings == nil {
			previewSettings, err = uc.loadAppPreviewSettings(ctx, db, app)
			if err != nil {
				return hperrors.Wrap(err)
			}
		}

		cloneDBApps := commentEvent.previewDeployCloneDB || previewSettings.AutoCloneApps
		if commentEvent.previewDeployNoCloneDB {
			cloneDBApps = false
		}
		previewTask, err = uc.appPreviewService.CreateAppPreviewTask(app, &entity.TaskAppPreviewArgs{
			ParentApp:       entity.ObjectID{ID: app.ID},
			RepoRef:         repoRef,
			NoStart:         commentEvent.previewDeployNoStart,
			CustomSubdomain: commentEvent.previewDeploySubdomain,
			CloneDBApps:     cloneDBApps,
			Trigger: &entity.AppDeploymentTrigger{
				Source:   base.DeploymentTriggerSourceRepoWebhook,
				SourceID: webhookID,
				ChangeID: "pr-" + strconv.FormatInt(commentEvent.PRNumber, 10),
			},
		})
		if err != nil {
			return hperrors.Wrap(err)
		}

		if !commentEvent.previewDeployNoWait && previewSettings.CreationDelay > 0 {
			previewTask.RunAt = previewTask.RunAt.Add(previewSettings.CreationDelay.ToDuration())
		}

		err = uc.taskRepo.Upsert(ctx, db, previewTask,
			entity.TaskUpsertingConflictCols, entity.TaskUpsertingUpdateCols)
		if err != nil {
			return hperrors.Wrap(err)
		}

		return nil
	})
	if err != nil {
		return hperrors.Wrap(err)
	}
	if previewTask != nil {
		_ = uc.taskQueue.ScheduleTask(ctx, previewTask)
	}
	return nil
}

func (uc *UC) loadAppPreviewSettings(
	ctx context.Context,
	db database.IDB,
	app *entity.App,
) (previewSettings *entity.AppFeaturePreviewSettings, err error) {
	_, featureSettings, err := uc.appService.LoadAppWithFeatureSettings(ctx, db, app.ProjectID, app.ID,
		false, false)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	previewSettings = featureSettings.PreviewSettings
	if previewSettings == nil || !previewSettings.Enabled {
		return nil, hperrors.Wrap(hperrors.ErrFeatureDisabled).
			WithParam("Name", "app preview")
	}

	return previewSettings, nil
}

func (uc *UC) deleteAppPreview(
	ctx context.Context,
	app *entity.App,
	expectedRef string,
) error {
	if !app.IsChildApp() { // must be a preview app to be deleted
		return nil
	}
	deploymentSetting := app.GetSettingByType(base.SettingTypeAppDeployment)
	if deploymentSetting == nil {
		return nil
	}
	deploymentSettings, err := deploymentSetting.AsAppDeploymentSettings()
	if err != nil {
		return hperrors.Wrap(err)
	}
	if deploymentSettings.ActiveMethod != base.DeploymentMethodRepo ||
		deploymentSettings.RepoSource == nil || deploymentSettings.RepoSource.RepoRef != expectedRef {
		return nil
	}

	err = transaction.Execute(ctx, uc.db, func(db database.Tx) error {
		if err = uc.appService.DeleteApp(ctx, db, app); err != nil {
			return hperrors.Wrap(err)
		}
		return nil
	})
	if err != nil {
		return hperrors.Wrap(err)
	}
	return nil
}
