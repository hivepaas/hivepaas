package projectuc

import (
	"context"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/base"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/entity"
	"github.com/localpaas/localpaas/localpaas_app/infra/database"
	"github.com/localpaas/localpaas/localpaas_app/pkg/bunex"
	"github.com/localpaas/localpaas/localpaas_app/pkg/timeutil"
	"github.com/localpaas/localpaas/localpaas_app/pkg/transaction"
	"github.com/localpaas/localpaas/localpaas_app/usecase/projectuc/projectdto"
)

func (uc *UC) UpdateProjectStatus(
	ctx context.Context,
	auth *basedto.Auth,
	req *projectdto.UpdateProjectStatusReq,
) (*projectdto.UpdateProjectStatusResp, error) {
	err := transaction.Execute(ctx, uc.db, func(db database.Tx) error {
		projectData := &updateProjectData{}
		err := uc.loadProjectDataForUpdateStatus(ctx, db, req, projectData)
		if err != nil {
			return apperrors.Wrap(err)
		}
		if !projectData.HasChanges {
			return nil
		}

		persistingData := &persistingProjectData{}
		uc.preparePersistingProjectStatusUpdate(req, projectData, persistingData)

		project := projectData.Project
		var targetAppStatus base.AppStatus
		switch project.Status {
		case base.ProjectStatusActive:
			targetAppStatus = base.AppStatusActive
		case base.ProjectStatusDisabled:
			targetAppStatus = base.AppStatusDisabled
		case base.ProjectStatusDeleting:
			// Do nothing
		}

		// TODO: lock apps
		if targetAppStatus != "" {
			for _, app := range project.Apps {
				if app.Status == targetAppStatus {
					continue
				}
				oldAppStatus := app.Status
				app.Status = targetAppStatus
				app.UpdatedAt = project.UpdatedAt
				app.UpdateVer++
				persistingData.UpsertingApps = append(persistingData.UpsertingApps, app)

				// TODO: handle error when update a specific app status
				_ = uc.appService.OnAppStatusChanged(ctx, app, oldAppStatus)
			}
		}

		return uc.persistData(ctx, db, persistingData)
	})
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	return &projectdto.UpdateProjectStatusResp{}, nil
}

func (uc *UC) loadProjectDataForUpdateStatus(
	ctx context.Context,
	db database.IDB,
	req *projectdto.UpdateProjectStatusReq,
	data *updateProjectData,
) error {
	project, err := uc.projectRepo.GetByID(ctx, db, req.ID,
		bunex.SelectFor("UPDATE"),
		bunex.SelectRelation("Apps",
			bunex.SelectExcludeColumns(entity.AppDefaultExcludeColumns...),
		),
	)
	if err != nil {
		return apperrors.Wrap(err)
	}
	if project.UpdateVer != req.UpdateVer {
		return apperrors.Wrap(apperrors.ErrUpdateVerMismatched)
	}
	data.Project = project
	data.HasChanges = project.Status != req.Status

	return nil
}

func (uc *UC) preparePersistingProjectStatusUpdate(
	req *projectdto.UpdateProjectStatusReq,
	data *updateProjectData,
	persistingData *persistingProjectData,
) {
	timeNow := timeutil.NowUTC()
	project := data.Project
	project.UpdateVer++
	project.Status = req.Status
	project.UpdatedAt = timeNow

	persistingData.UpsertingProjects = append(persistingData.UpsertingProjects, project)
}
