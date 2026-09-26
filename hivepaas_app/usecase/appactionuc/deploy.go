package appactionuc

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/transaction"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/appdeploymentservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/appservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/appactionuc/appactiondto"
)

func (uc *UC) DeployApp(
	ctx context.Context,
	auth *basedto.Auth,
	req *appactiondto.DeployAppReq,
) (*appactiondto.DeployAppResp, error) {
	var deployment *entity.Deployment
	var deploymentTask *entity.Task
	err := transaction.Execute(ctx, uc.db, func(db database.Tx) error {
		app, err := uc.appService.LoadApp(ctx, db, req.ProjectID, req.AppID, true, true,
			bunex.SelectFor("UPDATE OF app"),
			bunex.SelectRelation("Project",
				bunex.SelectExcludeColumns(entity.ProjectDefaultExcludeColumns...),
			),
			bunex.SelectRelation("ProjectEnv"),
			bunex.SelectRelation("Settings",
				bunex.SelectWhere("setting.type = ?", base.SettingTypeAppDeployment),
			),
		)
		if err != nil {
			return hperrors.Wrap(err)
		}

		deploymentSetting := app.GetSettingByType(base.SettingTypeAppDeployment)
		if deploymentSetting == nil || !deploymentSetting.IsActive() {
			return hperrors.NewNotFound("App deployment settings").
				WithMsgLog("app deployment settings not found")
		}
		deploymentSettings, err := deploymentSetting.AsAppDeploymentSettings()
		if err != nil {
			return hperrors.Wrap(err)
		}

		// Create a deployment and a task for it
		deployment, deploymentTask, err = uc.appDeploymentService.CreateDeploymentAndTask(
			app, deploymentSettings,
			appdeploymentservice.DeploymentArgs{NoCache: req.NoCache},
		)
		if err != nil {
			return hperrors.Wrap(err)
		}
		// Set trigger for the deployment
		deployment.Trigger = &entity.AppDeploymentTrigger{
			Source:   base.DeploymentTriggerSourceAPI,
			SourceID: auth.User.ID,
			ChangeID: req.ChangeID,
		}

		persistingData := &appservice.PersistingAppData{}
		persistingData.UpsertingDeployments = append(persistingData.UpsertingDeployments, deployment)
		persistingData.UpsertingTasks = append(persistingData.UpsertingTasks, deploymentTask)

		err = uc.appService.PersistAppData(ctx, db, persistingData)
		if err != nil {
			return hperrors.Wrap(err)
		}

		// Inside the transaction, unlike restart and stop: this one writes a
		// deployment row, so a failed record can still take the change with it.
		return uc.recordAppAction(ctx, db, auth, app, base.AuditLogSourceAPIAction, "deploy", nil)
	})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	err = uc.taskQueue.ScheduleTask(ctx, deploymentTask)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	return &appactiondto.DeployAppResp{
		Data: &appactiondto.DeployAppDataResp{
			DeploymentID: deployment.ID,
			TaskID:       deploymentTask.ID,
		},
	}, nil
}
