package appserviceimpl

import (
	"context"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/base"
	"github.com/localpaas/localpaas/localpaas_app/entity"
	"github.com/localpaas/localpaas/localpaas_app/infra/database"
	"github.com/localpaas/localpaas/localpaas_app/pkg/bunex"
	"github.com/localpaas/localpaas/localpaas_app/pkg/entityutil"
)

func (s *service) LoadApps(
	ctx context.Context,
	db database.IDB,
	projectID string,
	appIDs []string,
	requireProjectActive, requireAppsActive bool,
	extraOpts ...bunex.SelectQueryOption,
) ([]*entity.App, error) {
	// NOTE: make sure to add SelectRelation("Project") into extraOpts
	apps, err := s.appRepo.ListByIDs(ctx, db, projectID, appIDs, extraOpts...)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	appMap := entityutil.SliceToIDMap(apps)
	for _, appID := range appIDs {
		if _, exists := appMap[appID]; !exists {
			return nil, apperrors.NewNotFound(apperrors.Fmt("App '%v'", appID))
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
	extraOpts ...bunex.SelectQueryOption,
) (*entity.App, error) {
	// NOTE: make sure to add SelectRelation("Project") into extraOpts
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
		return apperrors.New(apperrors.ErrProjectInactive).WithNTParam("Name", projectName)
	}
	if requireAppActive && app.Status != base.AppStatusActive {
		return apperrors.New(apperrors.ErrAppInactive).WithNTParam("Name", app.Name)
	}
	return nil
}
