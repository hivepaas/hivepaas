package projectuc

import (
	"context"
	"errors"
	"strings"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/entity"
	"github.com/localpaas/localpaas/localpaas_app/infra/database"
	"github.com/localpaas/localpaas/localpaas_app/pkg/bunex"
	"github.com/localpaas/localpaas/localpaas_app/pkg/timeutil"
	"github.com/localpaas/localpaas/localpaas_app/pkg/transaction"
	"github.com/localpaas/localpaas/localpaas_app/usecase/projectuc/projectdto"
)

func (uc *UC) UpdateProject(
	ctx context.Context,
	auth *basedto.Auth,
	req *projectdto.UpdateProjectReq,
) (*projectdto.UpdateProjectResp, error) {
	err := transaction.Execute(ctx, uc.db, func(db database.Tx) error {
		projectData := &updateProjectData{}
		err := uc.loadProjectDataForUpdate(ctx, db, req, projectData)
		if err != nil {
			return apperrors.Wrap(err)
		}

		persistingData := &persistingProjectData{}
		uc.preparePersistingProjectUpdate(req, projectData, persistingData)

		return uc.persistData(ctx, db, persistingData)
	})
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	return &projectdto.UpdateProjectResp{}, nil
}

type updateProjectData struct {
	Project *entity.Project
}

func (uc *UC) loadProjectDataForUpdate(
	ctx context.Context,
	db database.IDB,
	req *projectdto.UpdateProjectReq,
	data *updateProjectData,
) error {
	project, err := uc.projectRepo.GetByID(ctx, db, req.ID,
		bunex.SelectFor("UPDATE"),
	)
	if err != nil {
		return apperrors.Wrap(err)
	}
	if project.UpdateVer != req.UpdateVer {
		return apperrors.Wrap(apperrors.ErrUpdateVerMismatched)
	}
	data.Project = project

	// If name changes, need to verify it uniqueness
	if !strings.EqualFold(req.Name, project.Name) {
		conflictProject, err := uc.projectRepo.GetByName(ctx, db, req.Name, bunex.SelectColumns("id"))
		if err != nil && !errors.Is(err, apperrors.ErrNotFound) {
			return apperrors.Wrap(err)
		}
		if conflictProject != nil {
			return apperrors.NewAlreadyExist("Project").
				WithMsgLog("project name '%s' already exists", req.Name)
		}
	}

	// Validate project owner
	if req.Owner.ID != "" && req.Owner.ID != project.OwnerID {
		_, err = uc.userService.LoadUser(ctx, db, req.Owner.ID)
		if err != nil {
			return apperrors.Wrap(err)
		}
	}

	return nil
}

func (uc *UC) preparePersistingProjectUpdate(
	req *projectdto.UpdateProjectReq,
	data *updateProjectData,
	persistingData *persistingProjectData,
) {
	project := data.Project
	project.UpdateVer++
	timeNow := timeutil.NowUTC()

	uc.preparePersistingProjectBase(project, req.ProjectBaseReq, timeNow, persistingData)
	persistingData.ProjectsToDeleteTags = append(persistingData.ProjectsToDeleteTags, project.ID)
	uc.preparePersistingProjectTags(project, req.Tags, 0, persistingData)
}
