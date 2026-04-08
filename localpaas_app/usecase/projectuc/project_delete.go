package projectuc

import (
	"context"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/entity"
	"github.com/localpaas/localpaas/localpaas_app/infra/database"
	"github.com/localpaas/localpaas/localpaas_app/pkg/bunex"
	"github.com/localpaas/localpaas/localpaas_app/pkg/timeutil"
	"github.com/localpaas/localpaas/localpaas_app/pkg/transaction"
	"github.com/localpaas/localpaas/localpaas_app/usecase/projectuc/projectdto"
)

func (uc *UC) DeleteProject(
	ctx context.Context,
	auth *basedto.Auth,
	req *projectdto.DeleteProjectReq,
) (*projectdto.DeleteProjectResp, error) {
	err := transaction.Execute(ctx, uc.db, func(db database.Tx) error {
		projectData := &deleteProjectData{}
		err := uc.loadProjectDataForDelete(ctx, db, req, projectData)
		if err != nil {
			return apperrors.Wrap(err)
		}

		persistingData := &persistingProjectData{}
		uc.prepareDeletingProject(projectData, persistingData)

		err = uc.persistData(ctx, db, persistingData)
		if err != nil {
			return apperrors.Wrap(err)
		}

		// Remove project and its app in infra
		err = uc.projectService.DeleteProject(ctx, projectData.Project)
		if err != nil {
			return apperrors.Wrap(err)
		}

		return nil
	})
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	return &projectdto.DeleteProjectResp{}, nil
}

type deleteProjectData struct {
	Project *entity.Project
}

func (uc *UC) loadProjectDataForDelete(
	ctx context.Context,
	db database.IDB,
	req *projectdto.DeleteProjectReq,
	data *deleteProjectData,
) error {
	project, err := uc.projectRepo.GetByID(ctx, db, req.ProjectID,
		bunex.SelectFor("UPDATE OF project"),
		bunex.SelectRelation("Apps",
			bunex.SelectWhere("app.deleted_at IS NULL"),
		),
	)
	if err != nil {
		return apperrors.Wrap(err)
	}
	data.Project = project

	return nil
}

func (uc *UC) prepareDeletingProject(
	data *deleteProjectData,
	persistingData *persistingProjectData,
) {
	timeNow := timeutil.NowUTC()
	project := data.Project
	project.DeletedAt = timeNow
	persistingData.UpsertingProjects = append(persistingData.UpsertingProjects, project)

	for _, app := range project.Apps {
		app.DeletedAt = timeNow
		persistingData.UpsertingApps = append(persistingData.UpsertingApps, app)
	}
}
