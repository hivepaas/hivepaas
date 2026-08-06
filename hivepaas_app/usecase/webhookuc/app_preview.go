package webhookuc

import (
	"context"
	"strconv"

	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/transaction"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/apppreviewservice"
)

func (uc *UC) createAppPreview(
	ctx context.Context,
	app *entity.App,
	commentEvent *repoPRCommentEventData,
	repoRef string,
	webhookID string,
) (err error) {
	if app.IsChildApp() { // The app is already a preview app, skips it
		return nil
	}
	var createResp *apppreviewservice.CreatePreviewResp
	defer func() {
		if (err != nil || recover() != nil) && createResp != nil && createResp.OnCleanup != nil {
			_ = createResp.OnCleanup(gofn.Coalesce(err, apperrors.ErrPanic))
		}
	}()

	err = transaction.Execute(ctx, uc.db, func(db database.Tx) (err error) {
		// Creating a preview may take time, so we don't lock the parent app.
		// However, after creating, we check the app status again. If it's not active and valid,
		// the preview app will be deleted.

		previewSettings, cloneApps, err := uc.loadAppPreviewSettings(ctx, db, app, commentEvent)
		if err != nil {
			return apperrors.Wrap(err)
		}

		createResp, err = uc.appPreviewService.CreatePreview(ctx, db, &apppreviewservice.CreatePreviewReq{
			App:             app,
			RepoRef:         repoRef,
			NoStart:         commentEvent.previewDeployNoStart,
			CustomSubdomain: commentEvent.previewDeploySubdomain,
			CloneDBApps:     cloneApps,
			OnInitDeployment: func(deployment *entity.Deployment) error {
				deployment.Trigger = &entity.AppDeploymentTrigger{
					Source:   base.DeploymentTriggerSourceRepoWebhook,
					SourceID: webhookID,
					ChangeID: "pr-" + strconv.FormatInt(commentEvent.PRNumber, 10),
				}
				return nil
			},
			OnDeploymentTask: func(task *entity.Task) error {
				task.RunAt = timeutil.NowUTC()
				if !commentEvent.previewDeployNoWait {
					task.RunAt = task.RunAt.Add(previewSettings.CreationDelay.ToDuration())
				}
				return nil
			},
		})
		if err != nil {
			return apperrors.Wrap(err)
		}

		// Ensure app valid when we complete creating the preview
		err = uc.appService.EnsureAppActive(ctx, db, app, false, false)
		if err != nil {
			return apperrors.Wrap(err)
		}
		return nil
	})
	if err != nil {
		return apperrors.Wrap(err)
	}
	if createResp != nil && createResp.DeploymentTask != nil {
		_ = uc.taskQueue.ScheduleTask(ctx, createResp.DeploymentTask)
	}
	return nil
}

func (uc *UC) loadAppPreviewSettings(
	ctx context.Context,
	db database.IDB,
	app *entity.App,
	commentEvent *repoPRCommentEventData,
) (previewSettings *entity.AppFeaturePreviewSettings, cloneApps []*entity.App, err error) {
	app, featureSettings, err := uc.appService.LoadAppWithFeatureSettings(ctx, db, app.ProjectID, app.ID,
		false, false)
	if err != nil {
		return nil, nil, apperrors.Wrap(err)
	}
	previewSettings = featureSettings.PreviewSettings
	if previewSettings == nil || !previewSettings.Enabled {
		return nil, nil, apperrors.Wrap(apperrors.ErrFeatureDisabled).
			WithParam("Name", "app preview")
	}

	var cloningAppIDs []string
	if commentEvent.previewDeployCloneDB || (previewSettings.AutoCloneApps && !commentEvent.previewDeployNoCloneDB) {
		cloningAppIDs = previewSettings.AppsToClone.ToIDStringSlice()
	}

	cloningApps, err := uc.appService.LoadAppsSkipMissing(ctx, db, app.Project.ID, cloningAppIDs,
		true, false,
		bunex.SelectExcludeColumns(entity.AppDefaultExcludeColumns...),
		bunex.SelectRelation("Project",
			bunex.SelectExcludeColumns(entity.ProjectDefaultExcludeColumns...),
		),
		bunex.SelectRelation("ProjectEnv"),
	)
	if err != nil {
		return nil, nil, apperrors.Wrap(err)
	}

	return previewSettings, cloningApps, nil
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
		return apperrors.Wrap(err)
	}
	if deploymentSettings.ActiveMethod != base.DeploymentMethodRepo ||
		deploymentSettings.RepoSource == nil || deploymentSettings.RepoSource.RepoRef != expectedRef {
		return nil
	}

	err = transaction.Execute(ctx, uc.db, func(db database.Tx) error {
		if err = uc.appService.DeleteApp(ctx, db, app); err != nil {
			return apperrors.Wrap(err)
		}
		return nil
	})
	if err != nil {
		return apperrors.Wrap(err)
	}
	return nil
}
