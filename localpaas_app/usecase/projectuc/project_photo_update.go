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

func (uc *UC) UpdateProjectPhoto(
	ctx context.Context,
	auth *basedto.Auth,
	req *projectdto.UpdateProjectPhotoReq,
) (*projectdto.UpdateProjectPhotoResp, error) {
	err := transaction.Execute(ctx, uc.db, func(db database.Tx) error {
		profileData := &updateProjectPhotoData{}
		err := uc.loadProjectPhotoDataForUpdate(ctx, db, req, profileData)
		if err != nil {
			return apperrors.Wrap(err)
		}

		persistingData := &persistingProjectPhotoData{}
		uc.preparePersistingProjectPhotoData(req, profileData, persistingData)

		return uc.persistProjectPhotoData(ctx, db, persistingData)
	})
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	return &projectdto.UpdateProjectPhotoResp{}, nil
}

type updateProjectPhotoData struct {
	Project *entity.Project
}

type persistingProjectPhotoData struct {
	UpdatingProject *entity.Project
}

func (uc *UC) loadProjectPhotoDataForUpdate(
	ctx context.Context,
	db database.IDB,
	req *projectdto.UpdateProjectPhotoReq,
	data *updateProjectPhotoData,
) error {
	project, err := uc.projectRepo.GetByID(ctx, db, req.ID,
		bunex.SelectFor("UPDATE"),
		bunex.SelectExcludeColumns(entity.ProjectDefaultExcludeColumns...),
	)
	if err != nil {
		return apperrors.Wrap(err)
	}
	data.Project = project

	// Save photo to local disk and set photo field of the project
	if req.FileExt != "" && len(req.DataBytes) > 0 {
		err = uc.projectService.SaveProjectPhoto(ctx, project, req.DataBytes, req.FileExt)
		if err != nil {
			return apperrors.Wrap(err)
		}
	}

	return nil
}

func (uc *UC) preparePersistingProjectPhotoData(
	req *projectdto.UpdateProjectPhotoReq,
	data *updateProjectPhotoData,
	persistingData *persistingProjectPhotoData,
) {
	timeNow := timeutil.NowUTC()
	project := data.Project
	persistingData.UpdatingProject = project
	project.UpdatedAt = timeNow

	if req.Delete {
		project.Photo = ""
	}
}

func (uc *UC) persistProjectPhotoData(
	ctx context.Context,
	db database.IDB,
	persistingData *persistingProjectPhotoData,
) error {
	err := uc.projectRepo.Update(ctx, db, persistingData.UpdatingProject,
		bunex.UpdateColumns("updated_at", "photo"),
	)
	if err != nil {
		return apperrors.Wrap(err)
	}
	return nil
}
