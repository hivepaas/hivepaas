package systemappserviceimpl

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/appdeploymentservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/appservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/systemappservice"
)

func (s *service) RecordImage(ctx context.Context, db database.IDB, app *entity.App, image string) error {
	setting := app.GetSettingByType(base.SettingTypeAppDeployment)
	if setting == nil {
		return hperrors.Wrap(hperrors.ErrInternal).
			WithMsgLog("system app '%s' has no deployment settings", app.Key)
	}
	settings, err := setting.AsAppDeploymentSettings()
	if err != nil {
		return hperrors.Wrap(err)
	}
	if settings.ImageSource != nil && settings.ImageSource.Image == image {
		return nil
	}
	if settings.ImageSource == nil {
		settings.ImageSource = &entity.DeploymentImageSource{}
	}
	settings.ImageSource.Image = image
	return hperrors.Wrap(s.persistDeploymentSettings(ctx, db, setting, settings))
}

func (s *service) persistDeploymentSettings(
	ctx context.Context,
	db database.IDB,
	setting *entity.Setting,
	settings *entity.AppDeploymentSettings,
) error {
	if err := setting.SetData(settings); err != nil {
		return hperrors.Wrap(err)
	}
	setting.UpdateVer++
	setting.UpdatedAt = timeutil.NowUTC()
	return hperrors.Wrap(s.settingRepo.Upsert(ctx, db, setting,
		entity.SettingUpsertingConflictCols, entity.SettingUpsertingUpdateCols))
}

// Redeploy goes through a deployment rather than a service update because the
// deployment is what applies the deployment settings: each one sets the
// container's arguments from the command stored there, so arguments written
// straight into the service would be wiped by the next deploy of the app.
func (s *service) Redeploy(
	ctx context.Context,
	db database.IDB,
	req *systemappservice.RedeployReq,
) (*entity.Task, error) {
	app := req.App
	setting := app.GetSettingByType(base.SettingTypeAppDeployment)
	if setting == nil {
		return nil, hperrors.Wrap(hperrors.ErrInternal).
			WithMsgLog("system app '%s' has no deployment settings to deploy", app.Key)
	}
	settings, err := setting.AsAppDeploymentSettings()
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	if !req.Change(settings) {
		return nil, nil
	}
	if err = s.persistDeploymentSettings(ctx, db, setting, settings); err != nil {
		return nil, hperrors.Wrap(err)
	}

	deployment, task, err := s.deploymentService.CreateDeploymentAndTask(app, settings,
		appdeploymentservice.DeploymentArgs{})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	deployment.Trigger = &entity.AppDeploymentTrigger{
		Source:   base.DeploymentTriggerSourceAPI,
		SourceID: req.TriggerUserID,
	}
	err = s.appService.PersistAppData(ctx, db, &appservice.PersistingAppData{
		UpsertingDeployments: []*entity.Deployment{deployment},
		UpsertingTasks:       []*entity.Task{task},
	})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	return task, nil
}
