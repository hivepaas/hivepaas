package webhookuc

import (
	"context"
	"time"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/transaction"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/appservice"
)

func (uc *UC) createAppDeployment(
	ctx context.Context,
	app *entity.App,
	changeID string,
	webhookID string,
) error {
	persistingData := &appservice.PersistingAppData{}
	err := transaction.Execute(ctx, uc.db, func(db database.Tx) error {
		err := uc.ensureAppActive(ctx, db, app, false, true)
		if err != nil {
			return apperrors.New(err)
		}
		err = uc.createAppDeploymentByChangeID(ctx, db, app, changeID, webhookID, persistingData)
		if err != nil {
			return apperrors.New(err)
		}
		err = uc.appService.PersistAppData(ctx, db, persistingData)
		if err != nil {
			return apperrors.New(err)
		}
		return nil
	})
	if err == nil && len(persistingData.UpsertingTasks) > 0 {
		_ = uc.taskQueue.ScheduleTask(ctx, persistingData.UpsertingTasks...)
	}
	if err != nil {
		return apperrors.New(err)
	}
	return nil
}

func (uc *UC) createAppDeploymentByChangeID(
	ctx context.Context,
	db database.Tx,
	app *entity.App,
	changeID string,
	webhookID string,
	persistingData *appservice.PersistingAppData,
) error {
	hasDeployment, err := uc.hasAppDeploymentByChangeID(ctx, db, app, changeID)
	if err != nil {
		return apperrors.New(err)
	}
	if hasDeployment {
		return nil
	}

	deploymentSetting := app.GetSettingByType(base.SettingTypeAppDeployment)
	deploymentSettings, err := deploymentSetting.AsAppDeploymentSettings()
	if err != nil {
		return apperrors.New(err)
	}
	if deploymentSettings.RepoSource != nil && deploymentSettings.RepoSource.CommitHash != "" {
		deploymentSettings.RepoSource.CommitHash = ""
		deploymentSetting.MustSetData(deploymentSettings)
		deploymentSetting.UpdateVer++
		deploymentSetting.UpdatedAt = timeutil.NowUTC()
		persistingData.UpsertingSettings = append(persistingData.UpsertingSettings, deploymentSetting)
	}

	deployment, task, err := uc.appDeploymentService.CreateDeploymentAndTask(app, deploymentSettings)
	if err != nil {
		return apperrors.New(err)
	}
	// Override target commit hash
	deployment.Settings.RepoSource.CommitHash = changeID
	// Set trigger for the deployment
	deployment.Trigger = &entity.AppDeploymentTrigger{
		Source:   base.DeploymentTriggerSourceRepoWebhook,
		SourceID: webhookID,
		ChangeID: changeID,
	}

	persistingData.UpsertingDeployments = append(persistingData.UpsertingDeployments, deployment)
	persistingData.UpsertingTasks = append(persistingData.UpsertingTasks, task)
	return nil
}

func (uc *UC) getAppDeploymentByChangeID(
	ctx context.Context,
	db database.Tx,
	app *entity.App,
	changeID string,
) (*entity.Deployment, error) {
	if changeID == "" {
		return nil, nil
	}
	deployments, _, err := uc.deploymentRepo.List(ctx, db, app.ID, nil,
		bunex.SelectColumns("id"),
		bunex.SelectLimit(1),
		bunex.SelectWhere("deployment.created_at > ?", timeutil.NowUTC().Add(-time.Minute)),
		bunex.SelectWhere("deployment.trigger->>'source' = ?", base.DeploymentTriggerSourceRepoWebhook),
		bunex.SelectWhere("deployment.trigger->>'changeId' = ?", changeID),
	)
	if err != nil {
		return nil, apperrors.New(err)
	}
	if len(deployments) == 0 {
		return nil, nil
	}
	return deployments[0], nil
}

func (uc *UC) hasAppDeploymentByChangeID(
	ctx context.Context,
	db database.Tx,
	app *entity.App,
	changeID string,
) (bool, error) {
	deployment, err := uc.getAppDeploymentByChangeID(ctx, db, app, changeID)
	if err != nil {
		return false, apperrors.New(err)
	}
	return deployment != nil, nil
}

func (uc *UC) ensureAppActive(
	ctx context.Context,
	db database.Tx,
	app *entity.App,
	checkUpdateVer bool,
	lockApp bool,
) error {
	qryOpts := []bunex.SelectQueryOption{
		bunex.SelectColumns("id"),
		bunex.SelectWhere("app.status = ?", base.AppStatusActive),
	}
	if checkUpdateVer {
		qryOpts = append(qryOpts,
			bunex.SelectWhere("app.update_ver = ?", app.UpdateVer),
		)
	}
	if lockApp {
		qryOpts = append(qryOpts, bunex.SelectFor("UPDATE OF app"))
	}
	_, err := uc.appRepo.GetByID(ctx, db, "", app.ID, qryOpts...)
	if err != nil {
		return apperrors.New(err)
	}
	return nil
}
