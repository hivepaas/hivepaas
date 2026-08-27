package projectuc

import (
	"context"
	"path/filepath"

	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/fileutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/transaction"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/ulid"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/projectuc/projectdto"
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
			return hperrors.Wrap(err)
		}

		persistingData := &persistingProjectPhotoData{}
		err = uc.preparePersistingProjectPhoto(ctx, db, req.ProjectPhotoReq, profileData.Project, persistingData)
		if err != nil {
			return hperrors.Wrap(err)
		}

		return uc.persistProjectPhotoData(ctx, db, persistingData)
	})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	return &projectdto.UpdateProjectPhotoResp{}, nil
}

type updateProjectPhotoData struct {
	Project *entity.Project
}

type persistingProjectPhotoData struct {
	UpdatingProject          *entity.Project
	UpsertingBinObjects      []*entity.BinObject
	HardDeletingBinObjectIDs []string
}

func (uc *UC) loadProjectPhotoDataForUpdate(
	ctx context.Context,
	db database.IDB,
	req *projectdto.UpdateProjectPhotoReq,
	data *updateProjectPhotoData,
) error {
	project, err := uc.projectRepo.GetByID(ctx, db, req.ID,
		bunex.SelectFor("UPDATE OF project"),
		bunex.SelectExcludeColumns(entity.ProjectDefaultExcludeColumns...),
		bunex.SelectRelation("PhotoData"),
	)
	if err != nil {
		return hperrors.Wrap(err)
	}
	data.Project = project

	return nil
}

func (uc *UC) preparePersistingProjectPhoto(
	ctx context.Context,
	db database.IDB,
	req *projectdto.ProjectPhotoReq,
	project *entity.Project,
	persistingData *persistingProjectPhotoData,
) error {
	if !req.IsChanged() {
		return nil
	}
	timeNow := timeutil.NowUTC()
	photoData := project.PhotoData

	if photoData != nil && photoData.ID != "" {
		// Project photo may take a remarkable space, so we hard-delete it if unused
		projects, _, err := uc.projectRepo.List(ctx, db, nil,
			bunex.SelectWhere("project.photo = ?", photoData.ID),
			bunex.SelectWhere("project.id != ?", project.ID),
			bunex.SelectLimit(1),
			bunex.SelectColumns("id"),
		)
		if err != nil {
			return hperrors.Wrap(err)
		}
		if len(projects) == 0 {
			persistingData.HardDeletingBinObjectIDs = append(persistingData.HardDeletingBinObjectIDs, photoData.ID)
		}
	}

	if req.Delete {
		project.Photo = ""
	} else {
		photoData = &entity.BinObject{
			ID:          gofn.Must(ulid.NewStringULID()),
			Type:        base.BinObjectTypeObjectIcon,
			Status:      base.BinObjectStatusActive,
			Name:        req.FileName,
			ContentType: fileutil.TypeByExtension(filepath.Ext(req.FileName)),
			Data:        req.DataBytes,
			CreatedAt:   timeNow,
			UpdatedAt:   timeNow,
		}
		persistingData.UpsertingBinObjects = append(persistingData.UpsertingBinObjects, photoData)
		project.Photo = photoData.ID
	}

	project.UpdatedAt = timeNow
	persistingData.UpdatingProject = project
	return nil
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
		return hperrors.Wrap(err)
	}

	err = uc.binObjectRepo.UpsertMulti(ctx, db, persistingData.UpsertingBinObjects,
		entity.BinObjectUpsertingConflictCols, entity.BinObjectUpsertingUpdateCols)
	if err != nil {
		return hperrors.Wrap(err)
	}

	err = uc.binObjectRepo.DeleteByIDs(ctx, db, persistingData.HardDeletingBinObjectIDs,
		bunex.DeleteWithForceDelete())
	if err != nil {
		return hperrors.Wrap(err)
	}
	return nil
}
