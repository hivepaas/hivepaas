package projectuc

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/transaction"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/projectuc/projectdto"
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

		for _, app := range project.Apps {
			if targetAppStatus == "" {
				continue
			}
			// Run app update in a separate transaction to reduce lock time
			err = uc.appService.ExecuteInTx(ctx, app, true, func(db database.Tx) error {
				err := uc.appService.SetAppStatus(ctx, db, app, targetAppStatus, true)
				if err != nil {
					return apperrors.Wrap(err)
				}
				return nil
			})
			if err != nil {
				return apperrors.Wrap(err)
			}
			return nil
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
		bunex.SelectFor("UPDATE OF project"),
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
