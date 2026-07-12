package appserviceimpl

import (
	"context"
	"errors"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/entityutil"
)

func (s *service) LoadApps(
	ctx context.Context,
	db database.IDB,
	projectID string,
	appIDs []string,
	requireProjectActive, requireAppsActive bool,
	extraOpts ...bunex.SelectQueryOption, // NOTE: make sure to add SelectRelation("Project")
) ([]*entity.App, error) {
	apps, err := s.appRepo.ListByIDs(ctx, db, projectID, appIDs, extraOpts...)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	appMap := entityutil.SliceToIDMap(apps)
	for _, appID := range appIDs {
		if _, exists := appMap[appID]; !exists {
			return nil, apperrors.Wrap(apperrors.ErrAppNotFound).WithParam("Name", appID)
		}
	}

	for _, app := range apps {
		if err = s.validateAppStatus(app, requireProjectActive, requireAppsActive); err != nil {
			return nil, apperrors.Wrap(err)
		}
	}
	return apps, nil
}

func (s *service) LoadApp(
	ctx context.Context,
	db database.IDB,
	projectID, appID string,
	requireProjectActive, requireAppActive bool,
	extraOpts ...bunex.SelectQueryOption, // NOTE: make sure to add SelectRelation("Project")
) (*entity.App, error) {
	app, err := s.appRepo.GetByID(ctx, db, projectID, appID, extraOpts...)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}
	if err = s.validateAppStatus(app, requireProjectActive, requireAppActive); err != nil {
		return nil, apperrors.Wrap(err)
	}
	return app, nil
}

func (s *service) LoadAppByKey(
	ctx context.Context,
	db database.IDB,
	projectID, appKey string,
	requireProjectActive, requireAppActive bool,
	extraOpts ...bunex.SelectQueryOption,
) (*entity.App, error) {
	// NOTE: make sure to add SelectRelation("Project") into extraOpts
	app, err := s.appRepo.GetByKey(ctx, db, projectID, appKey, extraOpts...)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}
	if err = s.validateAppStatus(app, requireProjectActive, requireAppActive); err != nil {
		return nil, apperrors.Wrap(err)
	}
	return app, nil
}

func (s *service) validateAppStatus(
	app *entity.App,
	requireProjectActive, requireAppActive bool,
) error {
	if requireProjectActive && (app.Project == nil || app.Project.Status != base.ProjectStatusActive) {
		projectName := app.ProjectID
		if app.Project != nil {
			projectName = app.Project.Name
		}
		return apperrors.Wrap(apperrors.ErrProjectInactive).WithNTParam("Name", projectName)
	}
	if requireAppActive {
		if app.Status != base.AppStatusActive {
			return apperrors.Wrap(apperrors.ErrAppInactive).WithNTParam("Name", app.Name)
		}
		if app.ParentApp != nil && app.ParentApp.Status != base.AppStatusActive {
			return apperrors.Wrap(apperrors.ErrAppInactive).WithNTParam("Name", app.ParentApp.Name)
		}
	}
	return nil
}

func (s *service) LoadAppWithFeatureSettings(
	ctx context.Context,
	db database.IDB,
	projectID, appID string,
	requireProjectActive, requireAppActive bool,
	extraOpts ...bunex.SelectQueryOption, // NOTE: make sure to add SelectRelation("Project")
) (app *entity.App, featureSettings *entity.AppFeatureSettings, err error) {
	app, err = s.LoadApp(ctx, db, projectID, appID, requireProjectActive, requireAppActive, extraOpts...)
	if err != nil {
		return nil, nil, apperrors.Wrap(err)
	}

	featureSetting, err := s.settingRepo.GetSingle(ctx, db, app.GetObjectScope(),
		base.SettingTypeAppFeatures, true)
	if err != nil && !errors.Is(err, apperrors.ErrNotFound) {
		return nil, nil, apperrors.Wrap(err)
	}
	if featureSetting != nil {
		featureSettings = featureSetting.MustAsAppFeatureSettings()
	} else {
		featureSettings = &entity.AppFeatureSettings{}
		entity.InitAppFeatureSettingsDefault(featureSettings)
	}
	return app, featureSettings, nil
}
