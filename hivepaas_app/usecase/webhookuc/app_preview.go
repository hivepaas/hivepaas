package webhookuc

import (
	"context"
	"strconv"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
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
) error {
	if app.ParentID != "" { // The app is already a preview app, skips it
		return nil
	}
	var createResp *apppreviewservice.CreatePreviewResp
	err := transaction.Execute(ctx, uc.db, func(db database.Tx) (err error) {
		createResp, err = uc.appPreviewService.CreatePreview(ctx, db, &apppreviewservice.CreatePreviewReq{
			ProjectID:       app.ProjectID,
			AppID:           app.ID,
			RepoRef:         repoRef,
			NoStart:         commentEvent.previewDeployNoStart,
			CustomSubdomain: commentEvent.previewDeploySubdomain,
			OnInitDeployment: func(deployment *entity.Deployment) error {
				deployment.Trigger = &entity.AppDeploymentTrigger{
					Source:   base.DeploymentTriggerSourceRepoWebhook,
					SourceID: webhookID,
					ChangeID: "pr-" + strconv.FormatInt(commentEvent.PRNumber, 10),
				}
				return nil
			},
			OnDeploymentTask: func(task *entity.Task) error {
				if !commentEvent.previewDeployNoWait {
					task.RunAt = timeutil.NowUTC().Add(deployDelayDefault)
				}
				return nil
			},
		})
		if err != nil {
			return apperrors.New(err)
		}
		return nil
	})
	if createResp != nil && createResp.OnCleanup != nil {
		_ = createResp.OnCleanup(err)
	}
	if err != nil {
		return apperrors.New(err)
	}
	if createResp != nil && createResp.DeploymentTask != nil {
		_ = uc.taskQueue.ScheduleTask(ctx, createResp.DeploymentTask)
	}
	return nil
}

func (uc *UC) deleteAppPreview(
	ctx context.Context,
	app *entity.App,
	expectedRef string,
) error {
	if app.ParentID == "" { // must be a preview app to be deleted
		return nil
	}
	deploymentSetting := app.GetSettingByType(base.SettingTypeAppDeployment)
	if deploymentSetting == nil {
		return nil
	}
	deploymentSettings, err := deploymentSetting.AsAppDeploymentSettings()
	if err != nil {
		return apperrors.New(err)
	}
	if deploymentSettings.ActiveMethod != base.DeploymentMethodRepo ||
		deploymentSettings.RepoSource == nil || deploymentSettings.RepoSource.RepoRef != expectedRef {
		return nil
	}

	err = transaction.Execute(ctx, uc.db, func(db database.Tx) error {
		if err = uc.appService.DeleteApp(ctx, db, app); err != nil {
			return apperrors.New(err)
		}
		return nil
	})
	if err != nil {
		return apperrors.New(err)
	}
	return nil
}
