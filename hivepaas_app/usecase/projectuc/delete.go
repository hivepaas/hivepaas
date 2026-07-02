package projectuc

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/transaction"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/projectuc/projectdto"
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
			return apperrors.New(err)
		}

		persistingData := &persistingProjectData{}
		uc.prepareDeletingProject(projectData, persistingData)

		err = uc.persistData(ctx, db, persistingData)
		if err != nil {
			return apperrors.New(err)
		}

		// Remove project and its app in infra
		err = uc.projectService.DeleteProject(ctx, db, projectData.Project)
		if err != nil {
			return apperrors.New(err)
		}

		return nil
	})
	if err != nil {
		return nil, apperrors.New(err)
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
		bunex.SelectFor("UPDATE"),
		bunex.SelectRelation("Apps",
			bunex.SelectWhere("app.deleted_at IS NULL"),
		),
	)
	if err != nil {
		return apperrors.New(err)
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
