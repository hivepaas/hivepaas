package appservice

import (
	"context"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/base"
	"github.com/localpaas/localpaas/localpaas_app/entity"
	"github.com/localpaas/localpaas/localpaas_app/infra/database"
	"github.com/localpaas/localpaas/localpaas_app/pkg/bunex"
)

func (s *appService) LoadApp(
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
	if requireProjectActive && (app.Project == nil || app.Project.Status != base.ProjectStatusActive) {
		projectName := app.ProjectID
		if app.Project != nil {
			projectName = app.Project.Name
		}
		return nil, apperrors.New(apperrors.ErrProjectInactive).WithNTParam("Name", projectName)
	}
	if requireAppActive && app.Status != base.AppStatusActive {
		return nil, apperrors.New(apperrors.ErrAppInactive).WithNTParam("Name", app.Name)
	}
	return app, nil
}

func (s *appService) LoadAppByToken(
	ctx context.Context,
	db database.IDB,
	appToken string,
	requireProjectActive, requireAppActive bool,
	extraOpts ...bunex.SelectQueryOption,
) (*entity.App, error) {
	// NOTE: make sure to add SelectRelation("Project") into extraOpts
	app, err := s.appRepo.GetByToken(ctx, db, appToken, extraOpts...)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}
	if requireProjectActive && (app.Project == nil || app.Project.Status != base.ProjectStatusActive) {
		projectName := app.ProjectID
		if app.Project != nil {
			projectName = app.Project.Name
		}
		return nil, apperrors.New(apperrors.ErrProjectInactive).WithNTParam("Name", projectName)
	}
	if requireAppActive && app.Status != base.AppStatusActive {
		return nil, apperrors.New(apperrors.ErrAppInactive).WithNTParam("Name", app.Name)
	}
	return app, nil
}
